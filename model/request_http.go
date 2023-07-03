// Package model -----------------------------
// @file      : request_http.go
// @author    : 被测试耽误的大厨
// @contact   : 13383088061@163.com
// @time      : 2023/6/27 18:25
// -------------------------------------------
package model

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/constant"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/tools"
	"github.com/comcast/go-edgegrid/edgegrid"
	"github.com/hiyosi/hawk"
	"github.com/lixiangyun/go-ntlm"
	"github.com/lixiangyun/go-ntlm/messages"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type HttpApiSetup struct {
	ClientName          string `json:"client_name"`
	IsRedirects         int64  `json:"is_redirects"`   // 是否跟随重定向 0: 是   1：否
	RedirectsNum        int    `json:"redirects_num"`  // 重定向次数>= 1; 默认为3
	ReadTimeOut         int64  `json:"read_time_out"`  // 请求读取超时时间
	WriteTimeOut        int64  `json:"write_time_out"` // 响应读取超时时间
	KeepAlive           bool   `json:"keep_alive"`
	MaxIdleConnDuration int64  `json:"max_idle_conn_duration"`
	MaxConnPerHost      int    `json:"max_conn_per_host"`
	UserAgent           bool   `json:"user_agent"`
	MaxConnWaitTimeout  int64  `json:"max_conn_wait_timeout"`
}

type Request struct {
	PreUrl       string               `json:"pre_url"`
	URL          string               `json:"url"`
	Method       string               `json:"method"` // 方法 GET/POST/PUT
	Debug        string               `json:"debug"`
	Parameter    []*VarForm           `json:"parameter"`
	Header       *Header              `json:"header"` // Headers
	Query        *Query               `json:"query"`
	Body         *Body                `json:"body"`
	Auth         *Auth                `json:"auth"`
	Cookie       *Cookie              `json:"cookie"`
	HttpApiSetup *HttpApiSetup        `json:"http_api_setup"`
	Assert       []*AssertionText     `json:"assert"` // 验证的方法(断言)
	Regex        []*RegularExpression `json:"regex"`  // 正则表达式
}

type Body struct {
	Mode      string     `json:"mode"`
	Raw       string     `json:"raw"`
	Parameter []*VarForm `json:"parameter"`
}

func (r Request) Send(debug string, debugMsg map[string]interface{}, requestCollection *mongo.Collection, globalVar *sync.Map) (bool, int64, uint64, float64, float64, string, time.Time, time.Time) {
	var (
		isSucceed       = true
		errCode         = constant.NoError
		receivedBytes   = float64(0)
		errMsg          = ""
		assertNum       = 0
		assertFailedNum = 0
	)

	if r.HttpApiSetup == nil {
		r.HttpApiSetup = new(HttpApiSetup)
	}

	resp, req, requestTime, sendBytes, err, str, startTime, endTime := r.Request()

	defer fasthttp.ReleaseResponse(resp) // 用完需要释放资源
	defer fasthttp.ReleaseRequest(req)
	var regex []map[string]interface{}
	if r.Regex != nil {
		for _, regular := range r.Regex {
			if regular.IsChecked != constant.Open {
				continue
			}
			reg := make(map[string]interface{})
			value := regular.Extract(resp, globalVar)
			if value == nil {
				continue
			}
			reg[regular.Var] = value
			regex = append(regex, reg)
		}
	}
	if err != nil {
		isSucceed = false
		errMsg = err.Error()
	}
	var assertionMsgList []AssertionMsg
	// 断言验证

	if r.Assert != nil {
		var assertionMsg = AssertionMsg{}
		var (
			code    = int64(10000)
			succeed = true
			msg     = ""
		)
		for _, v := range r.Assert {
			if v.IsChecked != constant.Open {
				continue
			}
			code, succeed, msg = v.VerifyAssertionText(resp)
			if succeed != true {
				errCode = code
				isSucceed = succeed
				errMsg = msg
				assertFailedNum++
			}
			assertionMsg.Code = code
			assertionMsg.IsSucceed = succeed
			assertionMsg.Msg = msg
			assertionMsgList = append(assertionMsgList, assertionMsg)
			assertNum++
		}
	}
	// 接收到的字节长度
	//contentLength = uint(resp.Header.ContentLength())

	receivedBytes = float64(resp.Header.ContentLength()) / 1024
	if receivedBytes <= 0 {
		receivedBytes = float64(len(resp.Body())) / 1024
	}
	// 开启debug模式后，将请求响应信息写入到mongodb中
	if debug == constant.All || debug == constant.OnlySuccess || debug == constant.OnlyError {
		responseTime := endTime.Format("2006-01-02 15:04:05")
		insertDebugMsg(regex, debugMsg, resp, req, requestTime, responseTime, receivedBytes, errMsg, debug, str, err, isSucceed, assertionMsgList, assertNum, assertFailedNum)
		if requestCollection != nil {
			Insert(requestCollection, debugMsg, middlewares.LocalIp)
		}
	}

	return isSucceed, errCode, requestTime, sendBytes, receivedBytes, errMsg, startTime, endTime
}

var (
	KeepAliveClient *fasthttp.Client
	once            sync.Once
)

func (r Request) Request() (resp *fasthttp.Response, req *fasthttp.Request, requestTime uint64, sendBytes float64, err error, str string, startTime, endTime time.Time) {
	var client *fasthttp.Client
	log.Logger.Debug("111111111111111")
	req = fasthttp.AcquireRequest()
	log.Logger.Debug("222222222222222")
	if r.HttpApiSetup.KeepAlive {
		newKeepAlive(r.HttpApiSetup, r.Auth)
		client = KeepAliveClient
		req.Header.Set("Connection", "keep-alive")
	} else {
		client = fastClient(r.HttpApiSetup, r.Auth)
	}

	// set method
	req.Header.SetMethod(r.Method)
	// set header
	r.Header.SetHeader(req)
	r.Cookie.SetCookie(req)
	url := r.URL
	urls := strings.Split(url, "//")
	if !strings.EqualFold(urls[0], constant.HTTP) && !strings.EqualFold(urls[0], constant.HTTPS) {
		url = constant.HTTP + "//" + url

	}

	urlQuery := req.URI().QueryArgs()

	if r.Query.Parameter != nil {
		for _, v := range r.Query.Parameter {
			if v.IsChecked != constant.Open {
				continue
			}
			if !strings.Contains(url, v.Key) {
				by := v.ValueToByte()
				urlQuery.AddBytesV(v.Key, by)
				url = url + fmt.Sprintf("&%s=%s", v.Key, string(v.ValueToByte()))
			}
		}
	}
	// set url
	req.SetRequestURI(url)
	// set body
	str = r.Body.SetBody(req)

	// set auth
	r.Auth.SetAuth(req)

	resp = fasthttp.AcquireResponse()
	startTime = time.Now()
	// 发送请求
	if r.HttpApiSetup.IsRedirects == 0 {
		err = client.DoRedirects(req, resp, r.HttpApiSetup.RedirectsNum)
	} else {
		err = client.Do(req, resp)
	}
	//err = client.Do(req, resp)
	endTime = time.Now()
	requestTime = uint64(time.Since(startTime))
	sendBytes = float64(req.Header.ContentLength()) / 1024
	if sendBytes <= 0 {
		sendBytes = float64(len(req.Body())) / 1024
	}
	return
}

// ReplaceQueryParameterizes 替换query中的变量
func (r *Request) ReplaceQueryParameterizes(globalVar *sync.Map) {

	// 将全局函数等，添加到api请求中
	if globalVar == nil {
		return
	}
	r.ReplaceUrl(globalVar)
	r.ReplaceBodyVarForm(globalVar)
	r.ReplaceQueryVarForm(globalVar)
	r.ReplaceHeaderVarForm(globalVar)
	r.ReplaceCookieVarForm(globalVar)
	r.ReplaceAuthVarForm(globalVar)
	r.ReplaceAssertionVarForm(globalVar)

}

func (r *Request) ReplaceUrl(globalVar *sync.Map) {
	urls := tools.FindAllDestStr(r.URL, "{{(.*?)}}")
	if urls == nil {
		return
	}
	for _, v := range urls {
		if len(v) < 2 {
			continue
		}
		realVar := tools.ParsFunc(v[1])
		if realVar != v[1] {
			r.URL = strings.Replace(r.URL, v[0], realVar, -1)
			continue
		}

		if globalVar == nil {
			continue
		}

		if value, ok := globalVar.Load(v[1]); ok {
			if value == nil {
				continue
			}
			switch fmt.Sprintf("%T", value) {
			case "int":
				value = fmt.Sprintf("%d", value)
			case "bool":
				value = fmt.Sprintf("%t", value)
			case "float64":
				value = fmt.Sprintf("%f", value)
			}
			r.URL = strings.Replace(r.URL, v[0], value.(string), -1)
		}

	}
}

func (r *Request) ReplaceBodyVarForm(globalVar *sync.Map) {
	if r.Body == nil {
		return
	}
	switch r.Body.Mode {
	case constant.NoneMode:
	case constant.FormMode:
		if r.Body.Parameter == nil || len(r.Body.Parameter) <= 0 {
			return
		}
		for _, queryVarForm := range r.Body.Parameter {
			keys := tools.FindAllDestStr(queryVarForm.Key, "{{(.*?)}}")
			if keys != nil {
				for _, v := range keys {
					if len(v) < 2 {
						continue
					}
					if value, ok := globalVar.Load(v[1]); ok {
						if value == nil {
							continue
						}
						switch fmt.Sprintf("%T", value) {
						case "int":
							value = fmt.Sprintf("%d", value)
						case "bool":
							value = fmt.Sprintf("%t", value)
						case "float64":
							value = fmt.Sprintf("%f", value)
						}
						queryVarForm.Key = strings.Replace(queryVarForm.Key, v[0], value.(string), -1)

					}
				}
			}
			if queryVarForm.Value == nil {
				continue
			}
			values := tools.FindAllDestStr(queryVarForm.Value.(string), "{{(.*?)}}")
			if values != nil {
				for _, v := range values {
					if len(v) < 2 {
						continue
					}
					realVar := tools.ParsFunc(v[1])
					if realVar != v[1] {
						queryVarForm.Value = strings.Replace(queryVarForm.Value.(string), v[0], realVar, -1)
						continue
					}

					if value, ok := globalVar.Load(v[1]); ok {
						if value == nil {
							continue
						}
						switch fmt.Sprintf("%T", value) {
						case "int":
							value = fmt.Sprintf("%d", value)
						case "bool":
							value = fmt.Sprintf("%t", value)
						case "float64":
							value = fmt.Sprintf("%f", value)
						}
						queryVarForm.Value = strings.Replace(queryVarForm.Value.(string), v[0], value.(string), -1)
					}
				}
			}
			queryVarForm.Conversion()
		}

	case constant.UrlencodeMode:
		if r.Body.Parameter == nil || len(r.Body.Parameter) <= 0 {
			return
		}
		for _, queryVarForm := range r.Body.Parameter {
			keys := tools.FindAllDestStr(queryVarForm.Key, "{{(.*?)}}")
			if keys != nil {
				for _, v := range keys {
					if len(v) < 2 {
						continue
					}
					if value, ok := globalVar.Load(v[1]); ok {
						if value == nil {
							continue
						}
						switch fmt.Sprintf("%T", value) {
						case "int":
							value = fmt.Sprintf("%d", value)
						case "bool":
							value = fmt.Sprintf("%t", value)
						case "float64":
							value = fmt.Sprintf("%f", value)
						}
						queryVarForm.Key = strings.Replace(queryVarForm.Key, v[0], value.(string), -1)
					}
				}
			}
			if queryVarForm.Value == nil {
				continue
			}
			values := tools.FindAllDestStr(queryVarForm.Value.(string), "{{(.*?)}}")
			if values != nil {
				for _, v := range values {
					if len(v) < 2 {
						continue
					}
					realVar := tools.ParsFunc(v[1])
					if realVar != v[1] {
						queryVarForm.Value = strings.Replace(queryVarForm.Value.(string), v[0], realVar, -1)
						continue
					}
					if value, ok := globalVar.Load(v[1]); ok {
						if value == nil {
							continue
						}
						switch fmt.Sprintf("%T", value) {
						case "int":
							value = fmt.Sprintf("%d", value)
						case "bool":
							value = fmt.Sprintf("%t", value)
						case "float64":
							value = fmt.Sprintf("%f", value)
						}
						queryVarForm.Value = strings.Replace(queryVarForm.Value.(string), v[0], value.(string), -1)
					}
				}
			}
			queryVarForm.Conversion()
		}
	default:
		bosys := tools.FindAllDestStr(r.Body.Raw, "{{(.*?)}}")
		if bosys == nil {
			return
		}
		for _, v := range bosys {
			if len(v) < 2 {
				continue
			}
			realVar := tools.ParsFunc(v[1])
			if realVar != v[1] {
				r.Body.Raw = strings.Replace(r.Body.Raw, v[0], realVar, -1)
				continue
			}

			if value, ok := globalVar.Load(v[1]); ok {
				if value == nil {
					continue
				}
				switch fmt.Sprintf("%T", value) {
				case "int":
					value = fmt.Sprintf("%d", value)
				case "bool":
					value = fmt.Sprintf("%t", value)
				case "float64":
					value = fmt.Sprintf("%f", value)
				}
				r.Body.Raw = strings.Replace(r.Body.Raw, v[0], value.(string), -1)
			}
		}
	}

}

func (r *Request) ReplaceHeaderVarForm(globalVar *sync.Map) {
	if r.Header == nil || r.Header.Parameter == nil {
		return
	}
	for _, queryVarForm := range r.Header.Parameter {
		queryParameterizes := tools.FindAllDestStr(queryVarForm.Key, "{{(.*?)}}")
		if queryParameterizes != nil {
			for _, v := range queryParameterizes {
				if len(v) < 2 {
					continue
				}
				if value, ok := globalVar.Load(v[1]); ok {
					if value == nil {
						continue
					}
					queryVarForm.Key = strings.Replace(queryVarForm.Key, v[0], value.(string), -1)
				}
			}
		}
		if queryVarForm.Value == nil {
			continue
		}
		queryParameterizes = tools.FindAllDestStr(queryVarForm.Value.(string), "{{(.*?)}}")
		if queryParameterizes == nil {
			continue
		}
		for _, v := range queryParameterizes {
			if len(v) < 2 {
				continue
			}
			realVar := tools.ParsFunc(v[1])
			if realVar != v[1] {
				queryVarForm.Value = strings.Replace(queryVarForm.Value.(string), v[0], realVar, -1)
				continue
			}
			if value, ok := globalVar.Load(v[1]); ok {
				if value == nil {
					continue
				}
				switch fmt.Sprintf("%T", value) {
				case "int":
					value = fmt.Sprintf("%d", value)
				case "bool":
					value = fmt.Sprintf("%t", value)
				case "float64":
					value = fmt.Sprintf("%f", value)
				}
				queryVarForm.Value = strings.Replace(queryVarForm.Value.(string), v[0], value.(string), -1)
			}

		}
		queryVarForm.Conversion()
	}
}

func (r *Request) ReplaceCookieVarForm(globalVar *sync.Map) {
	if r.Cookie == nil || r.Cookie.Parameter == nil {
		return
	}
	for _, queryVarForm := range r.Cookie.Parameter {
		queryParameterizes := tools.FindAllDestStr(queryVarForm.Key, "{{(.*?)}}")
		if queryParameterizes != nil {
			for _, v := range queryParameterizes {
				if len(v) < 2 {
					continue
				}
				if value, ok := globalVar.Load(v[1]); ok {
					if value == nil {
						continue
					}
					queryVarForm.Key = strings.Replace(queryVarForm.Key, v[0], value.(string), -1)
				}
			}
		}
		if queryVarForm.Value == nil {
			continue
		}
		queryParameterizes = tools.FindAllDestStr(queryVarForm.Value.(string), "{{(.*?)}}")
		if queryParameterizes == nil {
			continue
		}
		for _, v := range queryParameterizes {
			if len(v) < 2 {
				continue
			}
			realVar := tools.ParsFunc(v[1])
			if realVar != v[1] {
				queryVarForm.Value = strings.Replace(queryVarForm.Value.(string), v[0], realVar, -1)
				continue
			}
			if value, ok := globalVar.Load(v[1]); ok {
				if value == nil {
					continue
				}
				switch fmt.Sprintf("%T", value) {
				case "int":
					value = fmt.Sprintf("%d", value)
				case "bool":
					value = fmt.Sprintf("%t", value)
				case "float64":
					value = fmt.Sprintf("%f", value)
				default:
					by, _ := json.Marshal(value)
					if by != nil {
						value = string(by)
					}
				}
				queryVarForm.Value = strings.Replace(queryVarForm.Value.(string), v[0], value.(string), -1)
			}

		}
		queryVarForm.Conversion()
	}
}

func (r *Request) ReplaceQueryVarForm(globalVar *sync.Map) {
	if r.Query == nil || r.Query.Parameter == nil {
		return
	}
	for _, queryVarForm := range r.Query.Parameter {
		if queryVarForm.Value == nil {
			continue
		}
		keys := tools.FindAllDestStr(queryVarForm.Key, "{{(.*?)}}")
		if keys != nil {
			for _, v := range keys {
				if len(v) < 2 {
					continue
				}
				if value, ok := globalVar.Load(v[1]); ok {
					if value == nil {
						continue
					}
					queryVarForm.Key = strings.Replace(queryVarForm.Key, v[0], value.(string), -1)

				}
			}
		}

		values := tools.FindAllDestStr(queryVarForm.Value.(string), "{{(.*?)}}")
		if values == nil {
			continue
		}
		for _, v := range values {
			if len(v) < 2 {
				continue
			}
			realVar := tools.ParsFunc(v[1])
			if realVar != v[1] {
				queryVarForm.Value = strings.Replace(queryVarForm.Value.(string), v[0], realVar, -1)
				continue
			}
			if value, ok := globalVar.Load(v[1]); ok {
				if value == nil {
					continue
				}
				switch fmt.Sprintf("%T", value) {
				case "int":
					value = fmt.Sprintf("%d", value)
				case "bool":
					value = fmt.Sprintf("%t", value)
				case "float64":
					value = fmt.Sprintf("%f", value)
				}
				queryVarForm.Value = strings.Replace(queryVarForm.Value.(string), v[0], value.(string), -1)
			}
		}
		queryVarForm.Conversion()

	}

}

func (r *Request) ReplaceAuthVarForm(globalVar *sync.Map) {

	if r.Auth == nil {
		return
	}
	switch r.Auth.Type {
	case constant.Kv:
		if r.Auth.KV == nil || r.Auth.KV.Key == "" || r.Auth.KV.Value == nil {
			return
		}
		values := tools.FindAllDestStr(r.Auth.KV.Value.(string), "{{(.*?)}}")
		if values == nil {
			return
		}
		for _, v := range values {
			if len(v) < 2 {
				continue
			}
			realVar := tools.ParsFunc(v[1])
			if realVar != v[1] {
				r.Auth.KV.Value = strings.Replace(r.Auth.KV.Value.(string), v[0], realVar, -1)
				continue
			}
			if value, ok := globalVar.Load(v[1]); ok {
				if value == nil {
					continue
				}
				switch fmt.Sprintf("%T", value) {
				case "int":
					value = fmt.Sprintf("%d", value)
				case "bool":
					value = fmt.Sprintf("%t", value)
				case "float64":
					value = fmt.Sprintf("%f", value)
				}
				r.Auth.KV.Value = strings.Replace(r.Auth.KV.Value.(string), v[0], value.(string), -1)
			}
		}

	case constant.BEarer:
		if r.Auth.Bearer == nil || r.Auth.Bearer.Key == "" {
			return
		}
		values := tools.FindAllDestStr(r.Auth.Bearer.Key, "{{(.*?)}}")
		if values == nil {
			return
		}
		for _, v := range values {
			if len(v) < 2 {
				continue
			}
			realVar := tools.ParsFunc(v[1])
			if realVar != v[1] {
				r.Auth.Bearer.Key = strings.Replace(r.Auth.Bearer.Key, v[0], realVar, -1)
				continue
			}
			if value, ok := globalVar.Load(v[1]); ok {
				if value == nil {
					continue
				}
				switch fmt.Sprintf("%T", value) {
				case "int":
					value = fmt.Sprintf("%d", value)
				case "bool":
					value = fmt.Sprintf("%t", value)
				case "float64":
					value = fmt.Sprintf("%f", value)
				}
				r.Auth.Bearer.Key = strings.Replace(r.Auth.Bearer.Key, v[0], value.(string), -1)
			}
		}
	case constant.BAsic:
		if r.Auth.Basic != nil && r.Auth.Basic.UserName != "" {
			names := tools.FindAllDestStr(r.Auth.Basic.UserName, "{{(.*?)}}")
			if names != nil {
				for _, v := range names {
					if len(v) < 2 {
						continue
					}
					realVar := tools.ParsFunc(v[1])
					if realVar != v[1] {
						r.Auth.Basic.UserName = strings.Replace(r.Auth.Basic.UserName, v[0], realVar, -1)
						continue
					}
					if value, ok := globalVar.Load(v[1]); ok {
						if value == nil {
							continue
						}
						switch fmt.Sprintf("%T", value) {
						case "int":
							value = fmt.Sprintf("%d", value)
						case "bool":
							value = fmt.Sprintf("%t", value)
						case "float64":
							value = fmt.Sprintf("%f", value)
						}
						r.Auth.Basic.UserName = strings.Replace(r.Auth.Basic.UserName, v[0], value.(string), -1)

					}
				}
			}

		}

		if r.Auth.Basic == nil || r.Auth.Basic.Password == "" {
			return
		}
		passwords := tools.FindAllDestStr(r.Auth.Basic.Password, "{{(.*?)}}")
		if passwords == nil {
			return
		}
		for _, v := range passwords {
			if len(v) < 2 {
				continue
			}
			realVar := tools.ParsFunc(v[1])
			if realVar != v[1] {
				r.Auth.Basic.Password = strings.Replace(r.Auth.Basic.Password, v[0], realVar, -1)
				continue
			}
			if value, ok := globalVar.Load(v[1]); ok {
				if value == nil {
					continue
				}
				switch fmt.Sprintf("%T", value) {
				case "int":
					value = fmt.Sprintf("%d", value)
				case "bool":
					value = fmt.Sprintf("%t", value)
				case "float64":
					value = fmt.Sprintf("%f", value)
				}
				r.Auth.Basic.Password = strings.Replace(r.Auth.Basic.Password, v[0], value.(string), -1)
			}
		}
	}
}

func (r *Request) ReplaceAssertionVarForm(globalVar *sync.Map) {
	if r.Assert == nil || len(r.Assert) <= 0 {
		return
	}
	for _, assert := range r.Assert {
		if assert.Val == "" {
			continue
		}
		keys := tools.FindAllDestStr(assert.Var, "{{(.*?)}}")
		if keys != nil {
			for _, v := range keys {
				if len(v) < 2 {
					continue
				}
				if value, ok := globalVar.Load(v[1]); ok {
					if value == nil {
						continue
					}
					assert.Var = strings.Replace(assert.Val, v[0], value.(string), -1)

				}
			}
		}

		values := tools.FindAllDestStr(assert.Val, "{{(.*?)}}")
		if values == nil {
			continue
		}

		for _, v := range values {
			if len(v) < 2 {
				continue
			}
			realVar := tools.ParsFunc(v[1])
			if realVar != v[1] {
				assert.Val = strings.Replace(assert.Val, v[0], realVar, -1)
				continue
			}
			if value, ok := globalVar.Load(v[1]); ok {
				if value == nil {
					continue
				}
				switch fmt.Sprintf("%T", value) {
				case "int":
					value = fmt.Sprintf("%d", value)
				case "bool":
					value = fmt.Sprintf("%t", value)
				case "float64":
					value = fmt.Sprintf("%f", value)
				}
				assert.Val = strings.Replace(assert.Val, v[0], value.(string), -1)
			}
		}

	}
}

func (b *Body) SetBody(req *fasthttp.Request) string {
	if b == nil {
		return ""
	}
	switch b.Mode {
	case constant.NoneMode:
	case constant.FormMode:
		req.Header.SetContentType("multipart/form-data")
		// 新建一个缓冲，用于存放文件内容

		if b.Parameter == nil {
			b.Parameter = []*VarForm{}
		}

		bodyBuffer := &bytes.Buffer{}
		bodyWriter := multipart.NewWriter(bodyBuffer)
		contentType := bodyWriter.FormDataContentType()
		//var fileTypeList []string
		for _, value := range b.Parameter {

			if value.IsChecked != constant.Open {
				continue
			}
			if value.Key == "" {
				continue
			}

			switch value.Type {
			case constant.FileType:
				if value.FileBase64 == nil || len(value.FileBase64) < 1 {
					continue
				}
				for _, base64Str := range value.FileBase64 {
					by, fileType := tools.Base64DeEncode(base64Str, constant.FileType)
					log.Logger.Debug(fmt.Sprintf("机器ip:%s, fileType:    ", middlewares.LocalIp), fileType)
					if by == nil {
						continue
					}
					//fileWriter, err := bodyWriter.CreateFormFile(value.Key, value.Value.(string))
					h := make(textproto.MIMEHeader)
					h.Set("Content-Type", fileType)
					h.Set("Content-Disposition",
						fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
							value.Key, value.Value.(string)))
					fileWriter, err := bodyWriter.CreatePart(h)
					if err != nil {
						log.Logger.Error(fmt.Sprintf("机器ip:%s, CreateFormFile失败：%s ", middlewares.LocalIp, err.Error()))
						continue
					}
					file := bytes.NewReader(by)
					_, err = io.Copy(fileWriter, file)
					if err != nil {
						continue
					}
				}
			case constant.FileUrlType:
				val, ok := value.Value.(string)
				if !ok {
					continue
				}
				if strings.HasPrefix(val, "https://") || strings.HasPrefix(val, "http://") {
					strList := strings.Split(val, "/")
					if len(strList) < 1 {
						continue
					}
					fileTypeList := strings.Split(strList[len(strList)-1], ".")
					if len(fileTypeList) < 1 {
						continue
					}
					fc := &fasthttp.Client{}
					loadReq := fasthttp.AcquireRequest()
					defer loadReq.ConnectionClose()
					// set url
					loadReq.Header.SetMethod("GET")
					loadReq.SetRequestURI(val)
					loadResp := fasthttp.AcquireResponse()
					defer loadResp.ConnectionClose()
					if err := fc.Do(loadReq, loadResp); err != nil {
						log.Logger.Error(fmt.Sprintf("机器ip:%s, 下载body上传文件错误：", middlewares.LocalIp), err)
						continue
					}

					if loadResp.Body() == nil {
						continue
					}
					h := make(textproto.MIMEHeader)
					h.Set("Content-Type", fileTypeList[len(fileTypeList)-1])
					h.Set("Content-Disposition",
						fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
							value.Key, strList[len(strList)-1]))
					fileWriter, err := bodyWriter.CreatePart(h)
					if err != nil {
						log.Logger.Error(fmt.Sprintf("机器ip:%s, CreateFormFile失败：%s ", middlewares.LocalIp, err.Error()))
						continue
					}
					file := bytes.NewReader(loadResp.Body())
					_, err = io.Copy(fileWriter, file)
					if err != nil {
						continue
					}
				}

			default:
				filedWriter, err := bodyWriter.CreateFormField(value.Key)
				by := value.toByte()
				filed := bytes.NewReader(by)
				_, err = io.Copy(filedWriter, filed)
				if err != nil {
					log.Logger.Error(fmt.Sprintf("机器ip:%s, CreateFormFile失败： %s", middlewares.LocalIp, err.Error()))
					continue
				}
			}

		}
		bodyWriter.Close()
		req.Header.SetContentType(contentType)
		if bodyBuffer.Bytes() != nil && bodyBuffer.Len() != 68 {
			req.SetBody(bodyBuffer.Bytes())
		}
		return bodyBuffer.String()
	case constant.UrlencodeMode:

		req.Header.SetContentType("application/x-www-form-urlencoded")
		args := url.Values{}

		for _, value := range b.Parameter {
			if value.IsChecked != constant.Open || value.Key == "" || value.Value == nil {
				continue
			}
			args.Add(value.Key, value.Value.(string))

		}
		req.SetBodyString(args.Encode())
		return args.Encode()

	case constant.XmlMode:
		req.Header.SetContentType("application/xml")
		req.SetBodyString(b.Raw)
		return b.Raw
	case constant.JSMode:
		req.Header.SetContentType("application/javascript")
		req.SetBodyString(b.Raw)
		return b.Raw
	case constant.PlainMode:
		req.Header.SetContentType("text/plain")
		req.SetBodyString(b.Raw)
		return b.Raw
	case constant.HtmlMode:
		req.Header.SetContentType("text/html")
		req.SetBodyString(b.Raw)
		return b.Raw
	case constant.JsonMode:
		req.Header.SetContentType("application/json")
		req.SetBodyString(b.Raw)
		return b.Raw
	}
	return ""
}

type Header struct {
	Parameter []*VarForm `json:"parameter"`
}

func (header *Header) SetHeader(req *fasthttp.Request) {
	if header == nil || header.Parameter == nil {
		return
	}
	for _, v := range header.Parameter {
		if v.IsChecked != constant.Open || v.Value == nil {
			continue
		}
		if strings.EqualFold(v.Key, "content-type") {
			req.Header.SetContentType(v.Value.(string))
		}
		if strings.EqualFold(v.Key, "host") {
			req.SetHost(v.Value.(string))
			req.UseHostHeader = true
		}
		req.Header.Set(v.Key, v.Value.(string))
	}
}

func (cookie *Cookie) SetCookie(req *fasthttp.Request) {
	if cookie == nil || cookie.Parameter == nil {
		return
	}
	for _, v := range cookie.Parameter {
		if v.IsChecked != constant.Open || v.Value == nil || v.Key == "" {
			continue
		}
		req.Header.SetCookie(v.Key, v.Value.(string))
	}
}

type Query struct {
	Parameter []*VarForm `json:"parameter"`
}

type Cookie struct {
	Parameter []*VarForm `json:"parameter"`
}

type RegularExpression struct {
	IsChecked int         `json:"is_checked"` // 1 选中, 2未选
	Type      int         `json:"type"`       // 0 正则  1 json 2 header
	Var       string      `json:"var"`        // 变量
	Express   string      `json:"express"`    // 表达式
	Index     int         `json:"index"`      // 正则时提取第几个值
	Val       interface{} `json:"val"`        // 值
}

// Extract 提取response 中的值
func (re RegularExpression) Extract(resp *fasthttp.Response, globalVar *sync.Map) (value interface{}) {
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
		value = tools.FindAllDestStr(string(resp.Body()), re.Express)
		if value == nil || len(value.([][]string)) < 1 {
			value = ""
		} else {
			value = value.([][]string)[0][1]
		}
		globalVar.Store(name, value)
	case constant.JsonExtract:
		value = tools.JsonPath(string(resp.Body()), re.Express)
		globalVar.Store(name, value)
	case constant.HeaderExtract:
		if re.Express == "" {
			value = ""
			globalVar.Store(name, value)
			return
		}
		value = tools.MatchString(resp.Header.String(), re.Express, re.Index)
		globalVar.Store(name, value)
	case constant.CodeExtract:
		value = resp.StatusCode()
		globalVar.Store(name, value)
	}
	return
}

// VarForm 参数表
type VarForm struct {
	IsChecked   int64       `json:"is_checked" bson:"is_checked"`
	Type        string      `json:"type" bson:"type"`
	FileBase64  []string    `json:"fileBase64"`
	Key         string      `json:"key" bson:"key"`
	Value       interface{} `json:"value" bson:"value"`
	NotNull     int64       `json:"not_null" bson:"not_null"`
	Description string      `json:"description" bson:"description"`
	FieldType   string      `json:"field_type" bson:"field_type"`
}

func (vf *VarForm) VarFormTo(r *Api, globalVar *sync.Map) {
	for _, queryVarForm := range r.Request.Body.Parameter {
		keys := tools.FindAllDestStr(queryVarForm.Key, "{{(.*?)}}")
		if keys != nil {
			for _, v := range keys {
				if len(v) < 2 {
					continue
				}
				if value, ok := globalVar.Load(v[1]); ok {
					if value == nil {
						continue
					}
					queryVarForm.Key = strings.Replace(queryVarForm.Key, v[0], value.(string), -1)

				}
			}
		}
		if queryVarForm.Value == nil {
			continue
		}
		values := tools.FindAllDestStr(queryVarForm.Value.(string), "{{(.*?)}}")
		if values != nil {
			for _, v := range values {
				if len(v) < 2 {
					continue
				}
				realVar := tools.ParsFunc(v[1])
				if realVar != v[1] {
					queryVarForm.Value = strings.Replace(queryVarForm.Value.(string), v[0], realVar, -1)
					continue
				}

				if value, ok := globalVar.Load(v[1]); ok {
					if value == nil {
						continue
					}
					queryVarForm.Value = strings.Replace(queryVarForm.Value.(string), v[0], value.(string), -1)
				}
			}
		}
		queryVarForm.Conversion()
	}
}

type KV struct {
	Key   string      `json:"key" bson:"key"`
	Value interface{} `json:"value" bson:"value"`
}

type PlanKv struct {
	IsCheck int32  `json:"is_check"`
	Var     string `json:"Var"`
	Val     string `json:"Val"`
}

type Form struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type Bearer struct {
	Key string `json:"key" bson:"key"`
}

type Basic struct {
	UserName string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
}

type Digest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Realm     string `json:"realm"`
	Nonce     string `json:"nonce"`
	Algorithm string `json:"algorithm"`
	Qop       string `json:"qop"`
	Nc        string `json:"nc"`
	Cnonce    string `json:"cnonce"`
	Opaque    string `json:"opaque"`
}

type Hawk struct {
	AuthID             string `json:"authId"`
	AuthKey            string `json:"authKey"`
	Algorithm          string `json:"algorithm"`
	User               string `json:"user"`
	Nonce              string `json:"nonce"`
	ExtraData          string `json:"extraData"`
	App                string `json:"app"`
	Delegation         string `json:"delegation"`
	Timestamp          string `json:"timestamp"`
	IncludePayloadHash int    `json:"includePayloadHash"`
}

type AwsV4 struct {
	AccessKey          string `json:"accessKey"`
	SecretKey          string `json:"secretKey"`
	Region             string `json:"region"`
	Service            string `json:"service"`
	SessionToken       string `json:"sessionToken"`
	AddAuthDataToQuery int    `json:"addAuthDataToQuery"`
}

type Ntlm struct {
	Username            string `json:"username"`
	Password            string `json:"password"`
	Domain              string `json:"domain"`
	Workstation         string `json:"workstation"`
	DisableRetryRequest int    `json:"disableRetryRequest"`
}

type Edgegrid struct {
	AccessToken   string `json:"accessToken"`
	ClientToken   string `json:"clientToken"`
	ClientSecret  string `json:"clientSecret"`
	Nonce         string `json:"nonce"`
	Timestamp     string `json:"timestamp"`
	BaseURi       string `json:"baseURi"`
	HeadersToSign string `json:"headersToSign"`
}

type Oauth1 struct {
	ConsumerKey          string `json:"consumerKey"`
	ConsumerSecret       string `json:"consumerSecret"`
	SignatureMethod      string `json:"signatureMethod"`
	AddEmptyParamsToSign int    `json:"addEmptyParamsToSign"`
	IncludeBodyHash      int    `json:"includeBodyHash"`
	AddParamsToHeader    int    `json:"addParamsToHeader"`
	Realm                string `json:"realm"`
	Version              string `json:"version"`
	Nonce                string `json:"nonce"`
	Timestamp            string `json:"timestamp"`
	Verifier             string `json:"verifier"`
	Callback             string `json:"callback"`
	TokenSecret          string `json:"tokenSecret"`
	Token                string `json:"token"`
}
type Auth struct {
	Type          string    `json:"type" bson:"type"`
	Bidirectional *TLS      `json:"bidirectional"`
	KV            *KV       `json:"kv" bson:"kv"`
	Bearer        *Bearer   `json:"bearer" bson:"bearer"`
	Basic         *Basic    `json:"basic" bson:"basic"`
	Digest        *Digest   `json:"digest"`
	Hawk          *Hawk     `json:"hawk"`
	Awsv4         *AwsV4    `json:"awsv4"`
	Ntlm          *Ntlm     `json:"ntlm"`
	Edgegrid      *Edgegrid `json:"edgegrid"`
	Oauth1        *Oauth1   `json:"oauth1"`
}

type TLS struct {
	CaCert string `json:"ca_cert"`
}

type Token struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

type RequestData struct {
	Url           string `json:"url"`
	Method        string `json:"method"`
	Data          string `json:"data"`
	OauthCallback string `json:"oauth_callback"`
}

type Consumer struct {
}

func (auth *Auth) SetAuth(req *fasthttp.Request) {
	if auth == nil || auth.Type == constant.NoAuth || auth.Type == constant.Unidirectional || auth.Type == constant.Bidirectional {
		return
	}
	switch auth.Type {
	case constant.Kv:
		if auth.KV.Value != nil {
			req.Header.Add(auth.KV.Key, auth.KV.Value.(string))
		}

	case constant.BEarer:
		req.Header.Add("authorization", "Bearer "+auth.Bearer.Key)
	case constant.BAsic:
		pw := fmt.Sprintf("%s:%s", auth.Basic.UserName, auth.Basic.Password)
		req.Header.Add("Authorization", "Basic "+tools.Base64EncodeStd(pw))
	case constant.DigestType:
		encryption := tools.GetEncryption(auth.Digest.Algorithm)
		if encryption != nil {
			uri := string(req.URI().RequestURI())
			ha1 := ""
			ha2 := ""
			response := ""
			if auth.Digest.Cnonce == "" {
				auth.Digest.Cnonce = "apipost"
			}
			if auth.Digest.Nc == "" {
				auth.Digest.Nc = "00000001"
			}
			if strings.HasSuffix(auth.Digest.Algorithm, "-sess") {
				ha1 = encryption.HashFunc(encryption.HashFunc(auth.Digest.Username+":"+auth.Digest.Realm+":"+
					auth.Digest.Password) + ":" + auth.Digest.Nonce + ":" + auth.Digest.Cnonce)
			} else {
				ha1 = encryption.HashFunc(auth.Digest.Username + ":" + auth.Digest.Realm + ":" + auth.Digest.Password)
			}
			if auth.Digest.Qop != "auth-int" {
				ha2 = encryption.HashFunc(string(req.Header.Method()) + req.URI().String())
			} else {
				ha2 = encryption.HashFunc(string(req.Header.Method()) + uri + encryption.HashFunc(string(req.Body())))
			}
			if auth.Digest.Qop == "auth" || auth.Digest.Qop == "authn-int" {
				response = encryption.HashFunc(ha1 + ":" + auth.Digest.Nonce + ":" + auth.Digest.Nc +
					auth.Digest.Cnonce + ":" + auth.Digest.Qop + ":" + ha2)
			} else {
				response = encryption.HashFunc(ha1 + ":" + auth.Digest.Nonce + ":" + ha2)
			}
			digest := fmt.Sprintf("username=%s, realm=%s, nonce=%s, uri=%s, algorithm=%s, qop=%s, nc=%s, cnonce=%s, response=%s, opaque=%s",
				auth.Digest.Username, auth.Digest.Realm, auth.Digest.Nonce, uri, auth.Digest.Algorithm, auth.Digest.Qop,
				auth.Digest.Nc, auth.Digest.Cnonce, response, auth.Digest.Opaque)
			req.Header.Add("Authorization", digest)
		}
	case constant.HawkType:
		var alg hawk.Alg
		if strings.Contains(auth.Hawk.Algorithm, "SHA512") {
			alg = 2
		} else {
			alg = 1
		}
		credential := &hawk.Credential{
			ID:  auth.Hawk.AuthID,
			Key: auth.Hawk.AuthKey,
			Alg: alg,
		}
		timestamp, err := strconv.ParseInt(auth.Hawk.Timestamp, 10, 64)
		if err != nil {
			timestamp = time.Now().Unix()
		}
		option := &hawk.Option{
			TimeStamp: timestamp,
			Nonce:     auth.Hawk.Nonce,
			Ext:       auth.Hawk.ExtraData,
		}
		c := hawk.NewClient(credential, option)
		authorization, _ := c.Header(string(req.Header.Method()), string(req.Host())+string(req.Header.RequestURI()))
		req.Header.Add("Authorization", authorization)
	case constant.EdgegridType:
		reader := bytes.NewReader(req.Body())
		reqNew, err := http.NewRequest(string(req.Header.Method()), req.URI().String(), reader)
		if err != nil {
			return
		}
		params := edgegrid.NewAuthParams(reqNew, auth.Edgegrid.AccessToken, auth.Edgegrid.ClientToken, auth.Edgegrid.ClientSecret)
		authorization := edgegrid.Auth(params)
		req.Header.Add("Authorization", authorization)
	case constant.NtlmType:
		session, err := ntlm.CreateClientSession(ntlm.Version1, ntlm.ConnectionlessMode)
		if err != nil {
			return
		}
		session.SetUserInfo(auth.Ntlm.Username, auth.Ntlm.Password, auth.Ntlm.Domain)
		negotiate, err := session.GenerateNegotiateMessage()
		if err != nil {
			return
		}
		challenge, err := messages.ParseAuthenticateMessage(negotiate.Bytes, 2)
		if err != nil {
			return
		}
		req.Header.Add("Connection", "keep-alive")
		req.Header.Add("Authorization", challenge.String())
	case constant.Awsv4Type:
		signature := ""
		date := strconv.Itoa(int(time.Now().Month())) + strconv.Itoa(time.Now().Day())
		awsv := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/2022%s/%s/%s/aws4_request, SignedHeaders=content-length;content-type;host;x-amz-date;x-amz-security-token, Signature=%s",
			auth.Awsv4.AccessKey, date,
			auth.Awsv4.Region, auth.Awsv4.Service, signature)
		currentTime := strconv.Itoa(time.Now().Hour()) + strconv.Itoa(time.Now().Minute()) + strconv.Itoa(time.Now().Second())
		req.Header.Add("X-Amz-Security-Token", auth.Awsv4.SessionToken)
		req.Header.Add("X-Amz-Date", date+"T"+currentTime+"Z")
		req.Header.Add("Authorization", awsv)

	case constant.Oauth1Type:
	}
}
func (v *VarForm) ValueToByte() (by []byte) {
	if v.Value == nil {
		return
	}
	switch v.Type {
	case constant.StringType:
		by = []byte(v.Value.(string))
	case constant.TextType:
		by = []byte(v.Value.(string))
	case constant.ObjectType:
		by = []byte(v.Value.(string))
	case constant.ArrayType:
		by = []byte(v.Value.(string))
	case constant.NumberType:
		bytesBuffer := bytes.NewBuffer([]byte{})
		_ = binary.Write(bytesBuffer, binary.BigEndian, v.Value.(int))
		by = bytesBuffer.Bytes()
	case constant.IntegerType:
		bytesBuffer := bytes.NewBuffer([]byte{})
		_ = binary.Write(bytesBuffer, binary.BigEndian, v.Value.(int))
		by = bytesBuffer.Bytes()
	case constant.DoubleType:
		bytesBuffer := bytes.NewBuffer([]byte{})
		_ = binary.Write(bytesBuffer, binary.BigEndian, v.Value.(int64))
		by = bytesBuffer.Bytes()
	case constant.FileType:
		bits := math.Float64bits(v.Value.(float64))
		binary.LittleEndian.PutUint64(by, bits)
	case constant.BooleanType:
		buf := bytes.Buffer{}
		enc := gob.NewEncoder(&buf)
		_ = enc.Encode(v.Value.(bool))
		by = buf.Bytes()
	case constant.DateType:
		by = []byte(v.Value.(string))
	case constant.DateTimeType:
		by = []byte(v.Value.(string))
	case constant.TimeStampType:
		bytesBuffer := bytes.NewBuffer([]byte{})
		_ = binary.Write(bytesBuffer, binary.BigEndian, v.Value.(int64))
		by = bytesBuffer.Bytes()

	}
	return
}

func (v *VarForm) toByte() (by []byte) {
	if v.Value == nil {
		return
	}
	switch v.Type {
	case constant.StringType:
		by = []byte(v.Value.(string))
	case constant.TextType:
		by = []byte(v.Value.(string))
	case constant.ObjectType:
		by = []byte(v.Value.(string))
	case constant.ArrayType:
		by = []byte(v.Value.(string))
	case constant.NumberType:
		bytesBuffer := bytes.NewBuffer([]byte{})
		_ = binary.Write(bytesBuffer, binary.BigEndian, v.Value.(int))
		by = bytesBuffer.Bytes()
	case constant.IntegerType:
		bytesBuffer := bytes.NewBuffer([]byte{})
		_ = binary.Write(bytesBuffer, binary.BigEndian, v.Value.(int))
		by = bytesBuffer.Bytes()
	case constant.DoubleType:
		bytesBuffer := bytes.NewBuffer([]byte{})
		_ = binary.Write(bytesBuffer, binary.BigEndian, v.Value.(int64))
		by = bytesBuffer.Bytes()
	case constant.FileType:
		bits := math.Float64bits(v.Value.(float64))
		binary.LittleEndian.PutUint64(by, bits)
	case constant.BooleanType:
		buf := bytes.Buffer{}
		enc := gob.NewEncoder(&buf)
		_ = enc.Encode(v.Value.(bool))
		by = buf.Bytes()
	case constant.DateType:
		by = []byte(v.Value.(string))
	case constant.DateTimeType:
		by = []byte(v.Value.(string))
	case constant.TimeStampType:
		bytesBuffer := bytes.NewBuffer([]byte{})
		_ = binary.Write(bytesBuffer, binary.BigEndian, v.Value.(int64))
		by = bytesBuffer.Bytes()

	}
	return
}

// Conversion 将string转换为其他类型
func (v *VarForm) Conversion() {
	if v.Value == nil {
		return
	}
	switch v.FieldType {
	case constant.StringType:
		v.Value = v.Value.(string)
		// 字符串类型不用转换
	case constant.TextType:
		v.Value = v.Value.(string)
		// 文本类型不用转换
	case constant.ObjectType:
		v.Value = v.Value.(string)
		// 对象不用转换
	case constant.ArrayType:
		v.Value = v.Value.(string)
		// 数组不用转换
	case constant.IntegerType:
		v.Value = v.Value.(int)
	case constant.NumberType:
		v.Value = v.Value.(int)
	case constant.FloatType:
		v.Value = v.Value.(float64)
	case constant.DoubleType:
		v.Value = v.Value.(float64)
	case constant.FileType:
		v.Value = v.Value.(string)
	case constant.DateType:
		v.Value = v.Value.(string)
	case constant.DateTimeType:
		v.Value = v.Value.(string)
	case constant.TimeStampType:
		v.Value = v.Value.(int64)
	case constant.BooleanType:
		v.Value = v.Value.(bool)
	}
}

func insertDebugMsg(regex []map[string]interface{}, debugMsg map[string]interface{}, resp *fasthttp.Response, req *fasthttp.Request, requestTime uint64, responseTime string, receivedBytes float64, errMsg, debug, str string, err error, isSucceed bool, assertionMsgList []AssertionMsg, assertNum, assertFailedNum int) {
	switch debug {
	case constant.All:
		makeDebugMsg(regex, debugMsg, resp, req, requestTime, responseTime, receivedBytes, errMsg, str, err, isSucceed, assertionMsgList, assertNum, assertFailedNum)
	case constant.OnlySuccess:
		if isSucceed == true {
			makeDebugMsg(regex, debugMsg, resp, req, requestTime, responseTime, receivedBytes, errMsg, str, err, isSucceed, assertionMsgList, assertNum, assertFailedNum)
		}

	case constant.OnlyError:
		if isSucceed == false {
			makeDebugMsg(regex, debugMsg, resp, req, requestTime, responseTime, receivedBytes, errMsg, str, err, isSucceed, assertionMsgList, assertNum, assertFailedNum)
		}
	}
}

func makeDebugMsg(regex []map[string]interface{}, debugMsg map[string]interface{}, resp *fasthttp.Response, req *fasthttp.Request,
	requestTime uint64, responseTime string, receivedBytes float64, errMsg, str string, err error, isSucceed bool, assertionMsgList []AssertionMsg, assertNum, assertFailedNum int) {

	if req.Header.Method() != nil {
		debugMsg["method"] = string(req.Header.Method())
	}
	debugMsg["type"] = constant.RequestType
	debugMsg["request_time"] = requestTime / uint64(time.Millisecond)
	debugMsg["request_code"] = resp.StatusCode()
	debugMsg["request_header"] = req.Header.String()
	debugMsg["response_time"] = responseTime
	debugMsg["request_url"] = req.URI().String()
	if string(req.Body()) != "" {
		var errBody error
		debugMsg["request_body"], errBody = url.QueryUnescape(string(req.Body()))
		if errBody != nil {
			debugMsg["request_body"] = string(req.Body())
		}
	} else {
		debugMsg["request_body"] = str
	}
	if string(resp.Body()) == "" && errMsg != "" {
		debugMsg["response_body"] = errMsg
	}

	debugMsg["response_header"] = resp.Header.String()

	debugMsg["response_bytes"], _ = strconv.ParseFloat(fmt.Sprintf("%0.2f", receivedBytes), 64)
	if err != nil {
		debugMsg["response_body"] = err.Error()
	} else {
		debugMsg["response_body"] = string(resp.Body())
	}
	switch isSucceed {
	case false:
		debugMsg["status"] = constant.Failed
	case true:
		debugMsg["status"] = constant.Success
	}
	debugMsg["assertion"] = assertionMsgList
	debugMsg["assertion_num"] = assertNum
	debugMsg["assertion_failed_num"] = assertFailedNum
	debugMsg["regex"] = regex
}

// 获取fasthttp客户端
func fastClient(httpApiSetup *HttpApiSetup, auth *Auth) (fc *fasthttp.Client) {
	tr := &tls.Config{InsecureSkipVerify: true}
	if auth != nil || auth.Bidirectional != nil {
		switch auth.Type {
		case constant.Bidirectional:
			tr.InsecureSkipVerify = false
			if auth.Bidirectional.CaCert != "" {
				if strings.HasPrefix(auth.Bidirectional.CaCert, "https://") || strings.HasPrefix(auth.Bidirectional.CaCert, "http://") {
					client := &fasthttp.Client{}
					loadReq := fasthttp.AcquireRequest()
					defer loadReq.ConnectionClose()
					// set url
					loadReq.Header.SetMethod("GET")
					loadReq.SetRequestURI(auth.Bidirectional.CaCert)
					loadResp := fasthttp.AcquireResponse()
					defer loadResp.ConnectionClose()
					if err := client.Do(loadReq, loadResp); err != nil {
						log.Logger.Error(fmt.Sprintf("机器ip:%s, 下载crt文件失败：", middlewares.LocalIp), err)
					}
					if loadResp != nil && loadResp.Body() != nil {
						caCertPool := x509.NewCertPool()
						if caCertPool != nil {
							caCertPool.AppendCertsFromPEM(loadResp.Body())
							tr.ClientCAs = caCertPool
						}
					}
				}
			}
		case constant.Unidirectional:
			tr.InsecureSkipVerify = false
		}
	}
	fc = &fasthttp.Client{
		TLSConfig: tr,
	}
	if httpApiSetup.ClientName != "" {
		fc.Name = httpApiSetup.ClientName
	}
	if httpApiSetup.UserAgent {
		fc.NoDefaultUserAgentHeader = false
	}
	if httpApiSetup.MaxIdleConnDuration != 0 {
		fc.MaxIdleConnDuration = time.Duration(httpApiSetup.MaxIdleConnDuration) * time.Second
	} else {
		fc.MaxIdleConnDuration = time.Duration(0) * time.Second
	}
	if httpApiSetup.MaxConnPerHost != 0 {
		fc.MaxConnsPerHost = httpApiSetup.MaxConnPerHost
	}

	fc.MaxConnWaitTimeout = time.Duration(httpApiSetup.MaxConnWaitTimeout) * time.Second
	fc.WriteTimeout = time.Duration(httpApiSetup.WriteTimeOut) * time.Millisecond
	fc.ReadTimeout = time.Duration(httpApiSetup.ReadTimeOut) * time.Millisecond

	return fc
}

func newKeepAlive(httpApiSetup *HttpApiSetup, auth *Auth) {
	once.Do(func() {
		tr := &tls.Config{InsecureSkipVerify: true}
		if auth != nil && auth.Bidirectional != nil {
			switch auth.Type {
			case constant.Bidirectional:
				tr.InsecureSkipVerify = false
				if auth.Bidirectional.CaCert != "" {
					if strings.HasPrefix(auth.Bidirectional.CaCert, "https://") || strings.HasPrefix(auth.Bidirectional.CaCert, "http://") {
						client := &fasthttp.Client{}
						loadReq := fasthttp.AcquireRequest()
						defer loadReq.ConnectionClose()
						// set url
						loadReq.Header.SetMethod("GET")
						loadReq.SetRequestURI(auth.Bidirectional.CaCert)
						loadResp := fasthttp.AcquireResponse()
						defer loadResp.ConnectionClose()
						if err := client.Do(loadReq, loadResp); err != nil {
							log.Logger.Error(fmt.Sprintf("机器ip:%s, 下载crt文件失败：", middlewares.LocalIp), err)
						}
						if loadResp != nil && loadResp.Body() != nil {
							caCertPool := x509.NewCertPool()
							if caCertPool != nil {
								caCertPool.AppendCertsFromPEM(loadResp.Body())
								tr.ClientCAs = caCertPool
							}
						}
					}
				}
			case constant.Unidirectional:
				tr.InsecureSkipVerify = false
			}
		}
		KeepAliveClient = &fasthttp.Client{
			TLSConfig: tr,
		}
		if httpApiSetup.ClientName != "" {
			KeepAliveClient.Name = httpApiSetup.ClientName
		}
		if httpApiSetup.UserAgent {
			KeepAliveClient.NoDefaultUserAgentHeader = false
		}
		KeepAliveClient.MaxIdleConnDuration = time.Duration(httpApiSetup.MaxIdleConnDuration) * time.Second
		if httpApiSetup.MaxConnPerHost != 0 {
			KeepAliveClient.MaxConnsPerHost = httpApiSetup.MaxConnPerHost
		}

		KeepAliveClient.WriteTimeout = time.Duration(httpApiSetup.WriteTimeOut) * time.Millisecond
		KeepAliveClient.ReadTimeout = time.Duration(httpApiSetup.ReadTimeOut) * time.Millisecond
		KeepAliveClient.MaxConnWaitTimeout = time.Duration(httpApiSetup.MaxConnWaitTimeout) * time.Second
	})
}
