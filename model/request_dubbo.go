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

func (d DubboDetail) Send(debug string, debugMsg map[string]interface{}, mongoCollection *mongo.Collection, globalVariable *sync.Map) {
	parameterTypes, parameterValues := []string{}, []hessian.Object{}
	var err error
	var rpcServer common.RPCService
	soleKey := fmt.Sprintf("%s://%s/%s/%s", d.DubboProtocol, d.DubboConfig.RegistrationCenterAddress, d.ApiName, d.FunctionName)
	if s, ok := RpcServerMap.Load(soleKey); !ok {
		rpcServer, err = d.init(soleKey)
	} else {
		rpcServer = s
	}
	//rpcServer, err := d.init()
	for _, parame := range d.DubboParam {
		if parame.IsChecked != constant.Open {
			break
		}
		var val interface{}
		switch parame.ParamType {
		case constant.JavaInteger:
			val, err = strconv.Atoi(parame.Val)
			if err != nil {
				val = parame
				continue
			}
		case constant.JavaString:
			val = parame.Val
		case constant.JavaBoolean:
			switch parame.Val {
			case "true":
				val = true
			case "false":
				val = false
			default:
				val = parame.Val
			}
		case constant.JavaByte:

		case constant.JavaCharacter:
		case constant.JavaDouble:
			val, err = strconv.ParseFloat(parame.Val, 64)
			if err != nil {
				val = parame.Val
				continue
			}
		case constant.JavaFloat:
			val, err = strconv.ParseFloat(parame.Val, 64)
			if err != nil {
				val = parame.Val
				continue
			}
			val = float32(val.(float64))
		case constant.JavaLong:
			val, err = strconv.ParseInt(parame.Val, 10, 64)
			if err != nil {
				val = parame.Val
				continue
			}
		case constant.JavaMap:
		case constant.JavaList:
		default:
			val = parame.Val
		}
		parameterTypes = append(parameterTypes, parame.ParamType)
		parameterValues = append(parameterValues, val)

	}
	var resp interface{}
	var response []byte
	requestType, _ := json.Marshal(parameterTypes)
	debugMsg["request_parameter_type"] = string(requestType)
	requestBody, _ := json.Marshal(parameterValues)
	debugMsg["request_body"] = string(requestBody)
	if err != nil {
		debugMsg["status"] = false
		debugMsg["response_body"] = err.Error()
	} else {
		resp, err = rpcServer.(*generic.GenericService).Invoke(
			context.TODO(),
			d.FunctionName,
			parameterTypes,
			parameterValues, // 实参
		)
		if err != nil {
			debugMsg["status"] = false
			debugMsg["response_body"] = err.Error()
		}
		if resp != nil {
			response, _ = json.Marshal(resp)
			debugMsg["status"] = true
			debugMsg["response_body"] = string(response)

		}

	}
	var regex []map[string]interface{}
	if d.DubboRegex != nil {
		for _, regular := range d.DubboRegex {
			if regular.IsChecked != constant.Open {
				continue
			}
			reg := make(map[string]interface{})
			value := regular.Extract(string(response), globalVariable)
			if value == nil {
				continue
			}
			reg[regular.Var] = value
			regex = append(regex, reg)
		}
	}
	debugMsg["regex"] = regex

	var assertionMsgList []AssertionMsg
	// 断言验证

	if d.DubboAssert != nil {
		var (
			assertNum = 0
		)
		var assertionMsg = AssertionMsg{}
		var (
			code    = int64(10000)
			succeed = true
			msg     = ""
		)
		for _, v := range d.DubboAssert {
			if v.IsChecked != constant.Open {
				continue
			}
			code, succeed, msg = v.VerifyAssertionText(string(response))
			assertionMsg.Code = code
			assertionMsg.IsSucceed = succeed
			assertionMsg.Msg = msg
			assertionMsgList = append(assertionMsgList, assertionMsg)
			assertNum++
		}
	}
	debugMsg["assertion"] = assertionMsgList
	debugMsg["request_type"] = d.DubboProtocol
	Insert(mongoCollection, debugMsg, middlewares.LocalIp)
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

	//if err = rootConfig.Init(); err != nil {
	//	return
	//}

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
