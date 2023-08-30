// Package model -----------------------------
// @file      : request_dubbo.go
// @author    : 被测试耽误的大厨
// @contact   : 13383088061@163.com
// @time      : 2023/6/26 14:25
// -------------------------------------------
package model

import (
	"context"
	"dubbo.apache.org/dubbo-go/v3/common"
	dubboConfig "dubbo.apache.org/dubbo-go/v3/config"
	"dubbo.apache.org/dubbo-go/v3/config/generic"
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/constant"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/tools"
	hessian "github.com/apache/dubbo-go-hessian2"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"
	"strings"
	"sync"
)

type DubboDetail struct {
	TargetId string    `json:"target_id"`
	Uuid     uuid.UUID `json:"uuid"`
	Name     string    `json:"name"`
	TeamId   string    `json:"team_id"`
	Debug    string    `json:"debug"`

	DubboProtocol string `json:"dubbo_protocol"`
	ApiName       string `json:"api_name"`
	FunctionName  string `json:"function_name"`

	DubboParam     []DubboParam    `json:"dubbo_param"`
	DubboAssert    []DubboAssert   `json:"dubbo_assert"`
	DubboRegex     []DubboRegex    `json:"dubbo_regex"`
	DubboConfig    DubboConfig     `json:"dubbo_config"`
	Configuration  *Configuration  `json:"configuration"`   // 场景设置
	GlobalVariable *GlobalVariable `json:"global_variable"` // 全局变量
	DubboVariable  *GlobalVariable `json:"dubbo_variable"`
}

type DubboConfig struct {
	RegistrationCenterName    string `json:"registration_center_name"`
	RegistrationCenterAddress string `json:"registration_center_address"`
	Version                   string `json:"version"`
}

type DubboAssert struct {
	IsChecked    int    `json:"is_checked"`
	ResponseType int32  `json:"response_type"`
	Var          string `json:"var"`
	Compare      string `json:"compare"`
	Val          string `json:"val"`
	Index        int    `json:"index"` // 正则时提取第几个值
}

type DubboRegex struct {
	IsChecked int    `json:"is_checked"` // 1 选中, -1未选
	Type      int    `json:"type"`       // 0 正则  1 json
	Var       string `json:"var"`
	Express   string `json:"express"`
	Val       string `json:"val"`
	Index     int    `json:"index"` // 正则时提取第几个值
}

type DubboParam struct {
	IsChecked int32  `json:"is_checked"`
	ParamType string `json:"param_type"`
	Var       string `json:"var"`
	Val       string `json:"val"`
}

var RpcServerMap = new(sync.Map)

func (d DubboDetail) Send(debug string, debugMsg *DebugMsg, mongoCollection *mongo.Collection, globalVariable *sync.Map) {

	var (
		err             error
		regex           = &Regex{}
		success         string
		response        string
		rpcServer       common.RPCService
		assertNum       int
		parameterTypes  []string
		assertFailedNum int
		assert          = &Assert{}
		parameterValues []hessian.Object
	)
	d.DubboConfig.RegistrationCenterName = strings.TrimSpace(d.DubboConfig.RegistrationCenterName)
	d.DubboConfig.RegistrationCenterAddress = strings.TrimSpace(d.DubboConfig.RegistrationCenterAddress)
	d.ApiName = strings.TrimSpace(d.ApiName)
	d.DubboConfig.Version = strings.TrimSpace(d.DubboConfig.Version)
	soleKey := fmt.Sprintf("%s://%s/%s", d.DubboProtocol, d.DubboConfig.RegistrationCenterAddress, d.ApiName)

	if s, ok := RpcServerMap.Load(soleKey); !ok {
		rpcServer, err = d.init(soleKey)
	} else {
		rpcServer = s
	}
	parameterTypes, parameterValues = d.paramInit()

	// 发起反射调用
	if rpcServer != nil {
		success, response = d.invoke(err, rpcServer.(*generic.GenericService), parameterTypes, parameterValues)
	} else {
		success = constant.Failed
		if err != nil {
			response = err.Error()
		}

	}

	// 关联提取
	d.disExtract(response, globalVariable, regex)

	// 断言验证
	assertNum, assertFailedNum = d.disVerify(response, assert)

	requestType, _ := json.Marshal(parameterTypes)
	requestBody, _ := json.Marshal(parameterValues)

	if debug == constant.All {

		debugMsg.Regex = regex
		debugMsg.Assert = assert
		debugMsg.Status = success
		debugMsg.AssertNum = assertNum
		debugMsg.RequestUrl = d.ApiName
		debugMsg.Method = d.FunctionName
		debugMsg.RequestType = d.DubboProtocol
		debugMsg.RequestBody = string(requestBody)
		debugMsg.ResponseBody = response
		debugMsg.AssertFailedNum = assertFailedNum
		debugMsg.RequestParameterType = string(requestType)
		Insert(mongoCollection, debugMsg, middlewares.LocalIp)
	}
	if debug == constant.OnlySuccess && success == constant.Success {
		debugMsg.Regex = regex
		debugMsg.Assert = assert
		debugMsg.Status = success
		debugMsg.AssertNum = assertNum
		debugMsg.RequestUrl = d.ApiName
		debugMsg.Method = d.FunctionName
		debugMsg.RequestType = d.DubboProtocol
		debugMsg.RequestBody = string(requestBody)
		debugMsg.ResponseBody = response
		debugMsg.AssertFailedNum = assertFailedNum
		debugMsg.RequestParameterType = string(requestType)
		Insert(mongoCollection, debugMsg, middlewares.LocalIp)
	}
	if debug == constant.OnlyError && success == constant.Failed {
		debugMsg.Regex = regex
		debugMsg.Assert = assert
		debugMsg.Status = success
		debugMsg.AssertNum = assertNum
		debugMsg.RequestUrl = d.ApiName
		debugMsg.Method = d.FunctionName
		debugMsg.RequestType = d.DubboProtocol
		debugMsg.RequestBody = string(requestBody)
		debugMsg.ResponseBody = response
		debugMsg.AssertFailedNum = assertFailedNum
		debugMsg.RequestParameterType = string(requestType)
		Insert(mongoCollection, debugMsg, middlewares.LocalIp)
	}

}

func (d DubboDetail) init(soleKey string) (rpcServer common.RPCService, err error) {
	defer tools.DeferPanic("初始化dubbo配置失败")
	registryConfig := &dubboConfig.RegistryConfig{
		Protocol: d.DubboConfig.RegistrationCenterName,
		Address:  d.DubboConfig.RegistrationCenterAddress,
	}
	var zk string
	if d.DubboConfig.RegistrationCenterName == "zookeeper" {
		zk = "zk"
	} else {
		zk = d.DubboConfig.RegistrationCenterName
	}

	refConf := &dubboConfig.ReferenceConfig{
		InterfaceName:  d.ApiName, // 服务接口名，如：org.apache.dubbo.sample.UserProvider
		Cluster:        "failover",
		RegistryIDs:    []string{zk},          // 注册中心
		Protocol:       d.DubboProtocol,       // dubbo  或 tri（triple）  使用的协议
		Generic:        "true",                // true: 使用泛化调用；false: 不适用泛化调用
		Version:        d.DubboConfig.Version, // 版本号
		RequestTimeout: "3",
		Serialization:  "hessian2",
	}

	// 构造 Root 配置，引入注册中心模块
	rootConfig := dubboConfig.NewRootConfigBuilder().AddRegistry(zk, registryConfig).Build()
	if err = dubboConfig.Load(dubboConfig.WithRootConfig(rootConfig)); err != nil {
		return
	}

	// Reference 配置初始化，因为需要使用注册中心进行服务发现，需要传入经过配置的 rootConfig
	if err = refConf.Init(rootConfig); err != nil {
		return
	}

	if s, ok := RpcServerMap.Load(soleKey); !ok {
		refConf.GenericLoad(uuid.NewV4().String())
		//rpcServer = refConf.GetRPCService()

		rpcServer = refConf.GetRPCService()
		RpcServerMap.Store(soleKey, rpcServer)
	} else {
		rpcServer = s
	}
	return
}

// 发起反射调用
func (d DubboDetail) invoke(err error, rpcServer *generic.GenericService, parameterTypes []string, parameterValues []hessian.Object) (success, response string) {
	var resp interface{}
	if err != nil {
		success = constant.Failed
		response = err.Error()
		return
	} else {
		resp, err = rpcServer.Invoke(
			context.TODO(),
			d.FunctionName,  // 调用的方法
			parameterTypes,  // 参数类型
			parameterValues, // 实参
		)
		if resp != nil {
			switch fmt.Sprintf("%T", resp) {
			case constant.InterfaceMap:
				var fixedData map[string]interface{}
				fixedData = tools.FormatMap(resp.(map[interface{}]interface{}))
				by, _ := json.Marshal(fixedData)
				response = string(by)
			default:
				response = fmt.Sprint(resp)
			}
			success = constant.Success
		}
		if err != nil {
			success = constant.Failed
			response = err.Error()
		}

	}
	return
}

// 初始化参数
func (d DubboDetail) paramInit() (parameterTypes []string, parameterValues []hessian.Object) {
	if d.DubboParam == nil {
		return
	}
	var err error
	for _, param := range d.DubboParam {
		if param.IsChecked != constant.Open {
			break
		}
		var val interface{}
		switch param.ParamType {
		case constant.JavaInteger:
			val, err = strconv.Atoi(param.Val)
			if err != nil {
				val = param
				continue
			}
		case constant.JavaString:
			val = param.Val
		case constant.JavaBoolean:
			switch param.Val {
			case "true":
				val = true
			case "false":
				val = false
			default:
				val = param.Val
			}
		case constant.JavaByte:

		case constant.JavaCharacter:
		case constant.JavaDouble:
			val, err = strconv.ParseFloat(param.Val, 64)
			if err != nil {
				val = param.Val
				continue
			}
		case constant.JavaFloat:
			val, err = strconv.ParseFloat(param.Val, 64)
			if err != nil {
				val = param.Val
				continue
			}
			val = float32(val.(float64))
		case constant.JavaLong:
			val, err = strconv.ParseInt(param.Val, 10, 64)
			if err != nil {
				val = param.Val
				continue
			}
		case constant.JavaMap:
		case constant.JavaList:
		default:
			val = param.Val
		}
		parameterTypes = append(parameterTypes, param.ParamType)
		parameterValues = append(parameterValues, val)

	}
	return
}

// 处理关联提取
func (d DubboDetail) disExtract(response string, globalVariable *sync.Map, regex *Regex) {
	if d.DubboRegex == nil {
		return
	}
	for _, regular := range d.DubboRegex {
		if regular.IsChecked != constant.Open {
			continue
		}
		reg := new(Reg)
		value := regular.Extract(response, globalVariable)
		if value == nil {
			continue
		}
		reg.Key = regular.Var
		reg.Value = value
		regex.Regs = append(regex.Regs, reg)
	}
	return
}

// 处理断言
func (d DubboDetail) disVerify(response string, assert *Assert) (assertNum, assertFailedNum int) {
	if d.DubboAssert == nil {
		return
	}
	var assertionMsg = AssertionMsg{}
	var (
		code    = int64(10000)
		succeed = true
		msg     string
	)
	for _, v := range d.DubboAssert {
		if v.IsChecked != constant.Open {
			continue
		}
		code, succeed, msg = v.VerifyAssertionText(response)
		assertionMsg.Code = code
		assertionMsg.IsSucceed = succeed
		assertionMsg.Msg = msg
		assert.AssertionMsgs = append(assert.AssertionMsgs, assertionMsg)
		assertNum++
		if !succeed {
			assertFailedNum++
		}
	}
	return
}

// Extract 提取response 中的值
func (re DubboRegex) Extract(resp string, globalVar *sync.Map) (value interface{}) {
	re.Var = strings.TrimSpace(re.Var)
	name := tools.VariablesMatch(re.Var)
	if name == "" {
		return
	}
	re.Express = strings.TrimSpace(re.Express)
	keys := tools.FindAllDestStr(re.Express, "{{(.*?)}}")
	if keys != nil {
		for _, key := range keys {
			if len(key) < 2 {
				continue
			}
			realVar := tools.ParsFunc(key[1])
			if realVar != key[1] {
				re.Express = strings.Replace(re.Express, key[0], realVar, -1)
				continue
			}
			if v, ok := globalVar.Load(key[1]); ok {
				if v == nil {
					continue
				}
				re.Express = strings.Replace(re.Express, key[0], v.(string), -1)
			}
		}
	}
	switch re.Type {
	case constant.RegExtract:
		if re.Express == "" {
			value = ""
			globalVar.Store(name, value)
			return
		}
		value = tools.FindAllDestStr(resp, re.Express)
		if value == nil && len(value.([][]string)) < 1 {
			value = ""
		} else {
			value = value.([][]string)[0][1]
		}
		globalVar.Store(name, value)
	case constant.JsonExtract:

		value = tools.JsonPath(resp, re.Express)
		globalVar.Store(name, value)
	}
	return
}

// VerifyAssertionText 验证断言 文本断言
func (assertionText *DubboAssert) VerifyAssertionText(resp string) (code int64, ok bool, msg string) {
	assertionText.Var = strings.TrimSpace(assertionText.Var)
	assertionText.Val = strings.TrimSpace(assertionText.Val)
	switch assertionText.ResponseType {
	case constant.ResponseData:
		switch assertionText.Compare {
		case constant.Includes:
			if strings.Contains(resp, assertionText.Val) {
				return constant.NoError, true, "响应体中包含：" + assertionText.Val + " 断言: 成功！"
			} else {
				return constant.AssertError, false, "响应体中包含：" + assertionText.Val + " 断言: 失败！"
			}
		case constant.UNIncludes:
			if strings.Contains(resp, assertionText.Val) {
				return constant.AssertError, false, "响应体中不包含：" + assertionText.Val + " 断言: 失败！"
			} else {
				return constant.NoError, true, "响应体中不包含：" + assertionText.Val + " 断言: 成功！"
			}
		case constant.NULL:
			if resp == "" {
				return constant.NoError, true, "响应体为空， 断言: 成功！"
			} else {
				return constant.AssertError, false, "响应体为空， 断言: 失败！"
			}
		case constant.NotNULL:
			if resp == "" {
				return constant.AssertError, false, "响应体不为空， 断言: 失败！"
			} else {
				return constant.NoError, true, "响应体不为空， 断言: 成功！"
			}
		default:
			return constant.AssertError, false, "响应体断言条件不正确！"
		}
	}
	return constant.AssertError, false, "未选择被断言体！"
}
