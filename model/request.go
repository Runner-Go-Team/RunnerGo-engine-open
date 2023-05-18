package model

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/tools"
	"github.com/ThomsonReutersEikon/go-ntlm/ntlm"
	"github.com/comcast/go-edgegrid/edgegrid"
	"github.com/hiyosi/hawk"
	"github.com/lixiangyun/go-ntlm/messages"
	uuid "github.com/satori/go.uuid"
	"github.com/valyala/fasthttp"
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

// Api 请求数据
type Api struct {
	TargetId       string               `json:"target_id"`
	Uuid           uuid.UUID            `json:"uuid"`
	Name           string               `json:"name"`
	TeamId         string               `json:"team_id"`
	TargetType     string               `json:"target_type"` // api/webSocket/tcp/grpc
	Method         string               `json:"method"`      // 方法 GET/POST/PUT
	Request        Request              `json:"request"`
	Assert         []*AssertionText     `json:"assert"`          // 验证的方法(断言)
	Regex          []*RegularExpression `json:"regex"`           // 正则表达式
	Debug          string               `json:"debug"`           // 是否开启Debug模式
	Connection     int64                `json:"connection"`      // 0:websocket长连接
	Configuration  *Configuration       `json:"configuration"`   // 场景设置
	GlobalVariable *GlobalVariable      `json:"global_variable"` // 全局变量
	ApiVariable    *GlobalVariable      `json:"api_variable"`
	HttpApiSetup   *HttpApiSetup        `json:"http_api_setup"`
}

func (api *Api) GlobalToRequest() {
	if api.GlobalVariable.Cookie != nil && len(api.GlobalVariable.Cookie.Parameter) > 0 {
		if api.Request.Cookie == nil {
			api.Request.Cookie = new(Cookie)
		}
		if api.Request.Cookie.Parameter == nil {
			api.Request.Cookie.Parameter = []*VarForm{}
		}
		for _, parameter := range api.GlobalVariable.Cookie.Parameter {
			if parameter.IsChecked != Open {
				continue
			}
			var isExist bool
			for _, value := range api.Request.Cookie.Parameter {
				if value.IsChecked == Open && parameter.Key == value.Key && parameter.Value == value.Value {
					isExist = true
				}
			}
			if isExist {
				continue
			}
			api.Request.Cookie.Parameter = append(api.Request.Cookie.Parameter, parameter)
		}
	}
	if api.GlobalVariable.Header != nil && len(api.GlobalVariable.Header.Parameter) > 0 {
		if api.Request.Header == nil {
			api.Request.Header = new(Header)
		}
		if api.Request.Header.Parameter == nil {
			api.Request.Header.Parameter = []*VarForm{}
		}
		for _, parameter := range api.GlobalVariable.Header.Parameter {
			if parameter.IsChecked != Open {
				continue
			}
			var isExist bool
			for _, value := range api.Request.Header.Parameter {
				if value.IsChecked == Open && parameter.Key == value.Key && parameter.Value == parameter.Value {
					isExist = true
				}
			}
			if isExist {
				continue
			}
			api.Request.Header.Parameter = append(api.Request.Header.Parameter, parameter)

		}
	}

	if api.GlobalVariable.Assert != nil && len(api.GlobalVariable.Assert) > 0 {
		if api.Assert == nil {
			api.Assert = []*AssertionText{}
		}
		for _, parameter := range api.GlobalVariable.Assert {
			if parameter.IsChecked != Open {
				continue
			}
			var isExist bool
			for _, asser := range api.Assert {
				if asser.IsChecked == Open && parameter.ResponseType == asser.ResponseType && parameter.Compare == asser.Compare && parameter.Val == asser.Val && parameter.Var == asser.Var {
					isExist = true
				}
			}
			if isExist {
				continue
			}
			api.Assert = append(api.Assert, parameter)

		}
	}

}

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
	PreUrl    string     `json:"pre_url"`
	URL       string     `json:"url"`
	Parameter []*VarForm `json:"parameter"`
	Header    *Header    `json:"header"` // Headers
	Query     *Query     `json:"query"`
	Body      *Body      `json:"body"`
	Auth      *Auth      `json:"auth"`
	Cookie    *Cookie    `json:"cookie"`
}

type Body struct {
	Mode      string     `json:"mode"`
	Raw       string     `json:"raw"`
	Parameter []*VarForm `json:"parameter"`
}

func (b *Body) SetBody(req *fasthttp.Request) string {
	if b == nil {
		return ""
	}
	switch b.Mode {
	case NoneMode:
	case FormMode:
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

			if value.IsChecked != Open {
				continue
			}
			if value.Key == "" {
				continue
			}

			switch value.Type {
			case FileType:
				if value.FileBase64 == nil || len(value.FileBase64) < 1 {
					continue
				}
				for _, base64Str := range value.FileBase64 {
					by, fileType := tools.Base64DeEncode(base64Str, FileType)
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
			case FileUrlType:
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
	case UrlencodeMode:

		req.Header.SetContentType("application/x-www-form-urlencoded")
		args := url.Values{}

		for _, value := range b.Parameter {
			if value.IsChecked != Open || value.Key == "" || value.Value == nil {
				continue
			}
			args.Add(value.Key, value.Value.(string))

		}
		req.SetBodyString(args.Encode())
		return args.Encode()

	case XmlMode:
		req.Header.SetContentType("application/xml")
		req.SetBodyString(b.Raw)
		return b.Raw
	case JSMode:
		req.Header.SetContentType("application/javascript")
		req.SetBodyString(b.Raw)
		return b.Raw
	case PlainMode:
		req.Header.SetContentType("text/plain")
		req.SetBodyString(b.Raw)
		return b.Raw
	case HtmlMode:
		req.Header.SetContentType("text/html")
		req.SetBodyString(b.Raw)
		return b.Raw
	case JsonMode:
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
		if v.IsChecked != Open || v.Value == nil {
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
		if v.IsChecked != Open || v.Value == nil || v.Key == "" {
			continue
		}
		req.Header.SetCookie(v.Key, v.Value.(string))
	}
}

type Query struct {
	Parameter []*VarForm `json:"parameter"`
}

func (r *Api) SetQuery() {

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
	case RegExtract:
		if re.Express == "" {
			value = ""
			globalVar.Store(name, value)
			return
		}
		value = tools.FindAllDestStr(string(resp.Body()), re.Express)
		if value == nil && len(value.([][]string)) < 1 {
			value = ""
		}
		globalVar.Store(name, value)
	case JsonExtract:
		value = tools.JsonPath(string(resp.Body()), re.Express)
		globalVar.Store(name, value)
	case HeaderExtract:
		if re.Express == "" {
			value = ""
			globalVar.Store(name, value)
			return
		}
		value = tools.MatchString(resp.Header.String(), re.Express, re.Index)
		globalVar.Store(name, value)
	case CodeExtract:
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
	if auth == nil || auth.Type == NoAuth || auth.Type == Unidirectional || auth.Type == Bidirectional {
		return
	}
	switch auth.Type {
	case Kv:
		if auth.KV.Value != nil {
			req.Header.Add(auth.KV.Key, auth.KV.Value.(string))
		}

	case BEarer:
		req.Header.Add("authorization", "Bearer "+auth.Bearer.Key)
	case BAsic:
		pw := fmt.Sprintf("%s:%s", auth.Basic.UserName, auth.Basic.Password)
		req.Header.Add("Authorization", "Basic "+tools.Base64EncodeStd(pw))
	case DigestType:
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
	case HawkType:
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
	case EdgegridType:
		reader := bytes.NewReader(req.Body())
		reqNew, err := http.NewRequest(string(req.Header.Method()), req.URI().String(), reader)
		if err != nil {
			return
		}
		params := edgegrid.NewAuthParams(reqNew, auth.Edgegrid.AccessToken, auth.Edgegrid.ClientToken, auth.Edgegrid.ClientSecret)
		authorization := edgegrid.Auth(params)
		req.Header.Add("Authorization", authorization)
	case NtlmType:
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
	case Awsv4Type:
		signature := ""
		date := strconv.Itoa(int(time.Now().Month())) + strconv.Itoa(time.Now().Day())
		awsv := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/2022%s/%s/%s/aws4_request, SignedHeaders=content-length;content-type;host;x-amz-date;x-amz-security-token, Signature=%s",
			auth.Awsv4.AccessKey, date,
			auth.Awsv4.Region, auth.Awsv4.Service, signature)
		currentTime := strconv.Itoa(time.Now().Hour()) + strconv.Itoa(time.Now().Minute()) + strconv.Itoa(time.Now().Second())
		req.Header.Add("X-Amz-Security-Token", auth.Awsv4.SessionToken)
		req.Header.Add("X-Amz-Date", date+"T"+currentTime+"Z")
		req.Header.Add("Authorization", awsv)

	case Oauth1Type:

		//config := oauth1.Config{
		//	ConsumerKey:    auth.Oauth1.ConsumerKey,
		//	ConsumerSecret: auth.Oauth1.ConsumerSecret,
		//	CallbackURL:    req.URI().String(),
		//	Endpoint:       twitter.AuthorizeEndpoint,
		//	Realm:          auth.Oauth1.Realm,
		//}
		//
		//token := oauth1.NewToken(auth.Oauth1.Token, auth.Oauth1.Callback)
		//
		//authorization :=
		//	req.Header.Add("Authorization", authorization)
	}
}
func (v *VarForm) ValueToByte() (by []byte) {
	if v.Value == nil {
		return
	}
	switch v.Type {
	case StringType:
		by = []byte(v.Value.(string))
	case TextType:
		by = []byte(v.Value.(string))
	case ObjectType:
		by = []byte(v.Value.(string))
	case ArrayType:
		by = []byte(v.Value.(string))
	case NumberType:
		bytesBuffer := bytes.NewBuffer([]byte{})
		_ = binary.Write(bytesBuffer, binary.BigEndian, v.Value.(int))
		by = bytesBuffer.Bytes()
	case IntegerType:
		bytesBuffer := bytes.NewBuffer([]byte{})
		_ = binary.Write(bytesBuffer, binary.BigEndian, v.Value.(int))
		by = bytesBuffer.Bytes()
	case DoubleType:
		bytesBuffer := bytes.NewBuffer([]byte{})
		_ = binary.Write(bytesBuffer, binary.BigEndian, v.Value.(int64))
		by = bytesBuffer.Bytes()
	case FileType:
		bits := math.Float64bits(v.Value.(float64))
		binary.LittleEndian.PutUint64(by, bits)
	case BooleanType:
		buf := bytes.Buffer{}
		enc := gob.NewEncoder(&buf)
		_ = enc.Encode(v.Value.(bool))
		by = buf.Bytes()
	case DateType:
		by = []byte(v.Value.(string))
	case DateTimeType:
		by = []byte(v.Value.(string))
	case TimeStampType:
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
	case StringType:
		by = []byte(v.Value.(string))
	case TextType:
		by = []byte(v.Value.(string))
	case ObjectType:
		by = []byte(v.Value.(string))
	case ArrayType:
		by = []byte(v.Value.(string))
	case NumberType:
		bytesBuffer := bytes.NewBuffer([]byte{})
		_ = binary.Write(bytesBuffer, binary.BigEndian, v.Value.(int))
		by = bytesBuffer.Bytes()
	case IntegerType:
		bytesBuffer := bytes.NewBuffer([]byte{})
		_ = binary.Write(bytesBuffer, binary.BigEndian, v.Value.(int))
		by = bytesBuffer.Bytes()
	case DoubleType:
		bytesBuffer := bytes.NewBuffer([]byte{})
		_ = binary.Write(bytesBuffer, binary.BigEndian, v.Value.(int64))
		by = bytesBuffer.Bytes()
	case FileType:
		bits := math.Float64bits(v.Value.(float64))
		binary.LittleEndian.PutUint64(by, bits)
	case BooleanType:
		buf := bytes.Buffer{}
		enc := gob.NewEncoder(&buf)
		_ = enc.Encode(v.Value.(bool))
		by = buf.Bytes()
	case DateType:
		by = []byte(v.Value.(string))
	case DateTimeType:
		by = []byte(v.Value.(string))
	case TimeStampType:
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
	case StringType:
		v.Value = v.Value.(string)
		// 字符串类型不用转换
	case TextType:
		v.Value = v.Value.(string)
		// 文本类型不用转换
	case ObjectType:
		v.Value = v.Value.(string)
		// 对象不用转换
	case ArrayType:
		v.Value = v.Value.(string)
		// 数组不用转换
	case IntegerType:
		v.Value = v.Value.(int)
	case NumberType:
		v.Value = v.Value.(int)
	case FloatType:
		v.Value = v.Value.(float64)
	case DoubleType:
		v.Value = v.Value.(float64)
	case FileType:
		v.Value = v.Value.(string)
	case DateType:
		v.Value = v.Value.(string)
	case DateTimeType:
		v.Value = v.Value.(string)
	case TimeStampType:
		v.Value = v.Value.(int64)
	case BooleanType:
		v.Value = v.Value.(bool)
	}
}

// ReplaceQueryParameterizes 替换query中的变量
func (r *Api) ReplaceQueryParameterizes(globalVar *sync.Map) {
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

func (r *Api) AddAssertion() {
	if r.Configuration.SceneVariable == nil || r.Configuration.SceneVariable.Assert == nil || len(r.Configuration.SceneVariable.Assert) <= 0 {
		return
	}
	if r.Assert == nil {
		r.Assert = r.Configuration.SceneVariable.Assert
		return
	}
	by1, _ := json.Marshal(r.Assert)
	log.Logger.Debug("222222:     ", string(by1))
	for _, assert := range r.Configuration.SceneVariable.Assert {
		r.Assert = append(r.Assert, assert)
	}
	by, _ := json.Marshal(r.Configuration.SceneVariable.Assert)
	log.Logger.Debug("byLLLLL:     ", string(by))
}

func (r *Api) ReplaceUrl(globalVar *sync.Map) {
	urls := tools.FindAllDestStr(r.Request.URL, "{{(.*?)}}")
	if urls == nil {
		return
	}
	for _, v := range urls {
		if len(v) < 2 {
			continue
		}
		realVar := tools.ParsFunc(v[1])
		if realVar != v[1] {
			r.Request.URL = strings.Replace(r.Request.URL, v[0], realVar, -1)
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
			r.Request.URL = strings.Replace(r.Request.URL, v[0], value.(string), -1)
		}

	}
}

func (r *Api) ReplaceBodyVarForm(globalVar *sync.Map) {
	if r.Request.Body == nil {
		return
	}
	switch r.Request.Body.Mode {
	case NoneMode:
	case FormMode:
		if r.Request.Body.Parameter == nil || len(r.Request.Body.Parameter) <= 0 {
			return
		}
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

	case UrlencodeMode:
		if r.Request.Body.Parameter == nil || len(r.Request.Body.Parameter) <= 0 {
			return
		}
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
		bosys := tools.FindAllDestStr(r.Request.Body.Raw, "{{(.*?)}}")
		if bosys == nil {
			return
		}
		for _, v := range bosys {
			if len(v) < 2 {
				continue
			}
			realVar := tools.ParsFunc(v[1])
			if realVar != v[1] {
				r.Request.Body.Raw = strings.Replace(r.Request.Body.Raw, v[0], realVar, -1)
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
				r.Request.Body.Raw = strings.Replace(r.Request.Body.Raw, v[0], value.(string), -1)
			}
		}
	}

}

func (r *Api) ReplaceHeaderVarForm(globalVar *sync.Map) {
	if r.Request.Header == nil || r.Request.Header.Parameter == nil {
		return
	}
	for _, queryVarForm := range r.Request.Header.Parameter {
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

func (r *Api) ReplaceCookieVarForm(globalVar *sync.Map) {
	if r.Request.Cookie == nil || r.Request.Cookie.Parameter == nil {
		return
	}
	for _, queryVarForm := range r.Request.Cookie.Parameter {
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

func (r *Api) ReplaceQueryVarForm(globalVar *sync.Map) {
	if r.Request.Query == nil || r.Request.Query.Parameter == nil {
		return
	}
	for _, queryVarForm := range r.Request.Query.Parameter {
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

func (r *Api) ReplaceAuthVarForm(globalVar *sync.Map) {

	if r.Request.Auth == nil {
		return
	}
	switch r.Request.Auth.Type {
	case Kv:
		if r.Request.Auth.KV == nil || r.Request.Auth.KV.Key == "" || r.Request.Auth.KV.Value == nil {
			return
		}
		values := tools.FindAllDestStr(r.Request.Auth.KV.Value.(string), "{{(.*?)}}")
		if values == nil {
			return
		}
		for _, v := range values {
			if len(v) < 2 {
				continue
			}
			realVar := tools.ParsFunc(v[1])
			if realVar != v[1] {
				r.Request.Auth.KV.Value = strings.Replace(r.Request.Auth.KV.Value.(string), v[0], realVar, -1)
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
				r.Request.Auth.KV.Value = strings.Replace(r.Request.Auth.KV.Value.(string), v[0], value.(string), -1)
			}
		}

	case BEarer:
		if r.Request.Auth.Bearer == nil || r.Request.Auth.Bearer.Key == "" {
			return
		}
		values := tools.FindAllDestStr(r.Request.Auth.Bearer.Key, "{{(.*?)}}")
		if values == nil {
			return
		}
		for _, v := range values {
			if len(v) < 2 {
				continue
			}
			realVar := tools.ParsFunc(v[1])
			if realVar != v[1] {
				r.Request.Auth.Bearer.Key = strings.Replace(r.Request.Auth.Bearer.Key, v[0], realVar, -1)
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
				r.Request.Auth.Bearer.Key = strings.Replace(r.Request.Auth.Bearer.Key, v[0], value.(string), -1)
			}
		}
	case BAsic:
		if r.Request.Auth.Basic != nil && r.Request.Auth.Basic.UserName != "" {
			names := tools.FindAllDestStr(r.Request.Auth.Basic.UserName, "{{(.*?)}}")
			if names != nil {
				for _, v := range names {
					if len(v) < 2 {
						continue
					}
					realVar := tools.ParsFunc(v[1])
					if realVar != v[1] {
						r.Request.Auth.Basic.UserName = strings.Replace(r.Request.Auth.Basic.UserName, v[0], realVar, -1)
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
						r.Request.Auth.Basic.UserName = strings.Replace(r.Request.Auth.Basic.UserName, v[0], value.(string), -1)

					}
				}
			}

		}

		if r.Request.Auth.Basic == nil || r.Request.Auth.Basic.Password == "" {
			return
		}
		passwords := tools.FindAllDestStr(r.Request.Auth.Basic.Password, "{{(.*?)}}")
		if passwords == nil {
			return
		}
		for _, v := range passwords {
			if len(v) < 2 {
				continue
			}
			realVar := tools.ParsFunc(v[1])
			if realVar != v[1] {
				r.Request.Auth.Basic.Password = strings.Replace(r.Request.Auth.Basic.Password, v[0], realVar, -1)
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
				r.Request.Auth.Basic.Password = strings.Replace(r.Request.Auth.Basic.Password, v[0], value.(string), -1)
			}
		}
	}
}

func (r *Api) ReplaceAssertionVarForm(globalVar *sync.Map) {
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

// FindParameterizes 将请求中的变量全部放到一个map中
//func (r *Api) FindParameterizes() {
//	r.Request.URL = strings.TrimSpace(r.Request.URL)
//	urls := tools.FindAllDestStr(r.Request.URL, "{{(.*?)}}")
//
//	for _, name := range urls {
//
//		r.Parameters.Range(func(key, value any) bool {
//			return true
//		})
//		if _, ok := r.Parameters.Load(name[1]); !ok {
//			r.Parameters.Store(name[1], name[0])
//		}
//	}
//	r.findBodyParameters()
//	r.findQueryParameters()
//	r.findHeaderParameters()
//	r.findAuthParameters()
//}

// ReplaceParameters 将场景变量中的值赋值给，接口变量
//func (r *Api) ReplaceParameters(Variable []*KV) {
//	for _, v := range Variable {
//		r.Parameters.Store(v.Key, v.Value)
//	}
//}

// 将Query中的变量，都存储到接口变量中
//func (r *Api) findQueryParameters() {
//
//	if r.Request.Query == nil || r.Request.Query.Parameter == nil {
//		return
//	}
//	for _, varForm := range r.Request.Query.Parameter {
//		nameParameters := tools.FindAllDestStr(varForm.Key, "{{(.*?)}}")
//		for _, name := range nameParameters {
//			if _, ok := r.Parameters.Load(name[1]); !ok {
//				r.Parameters.Store(name[1], name[0])
//			}
//		}
//		if varForm.Value == nil {
//			continue
//		}
//		valueParameters := tools.FindAllDestStr(varForm.Value.(string), "{{(.*?)}}")
//		for _, value := range valueParameters {
//			if len(value) > 1 {
//				if _, ok := r.Parameters.Load(value[1]); !ok {
//					r.Parameters.Store(value[1], value[0])
//				}
//			}
//
//		}
//	}
//
//}

//func (r *Api) findBodyParameters() {
//	if r.Request.Body != nil {
//		switch r.Request.Body.Mode {
//		case NoneMode:
//		case FormMode:
//			if r.Request.Body.Parameter == nil {
//				return
//			}
//			for _, parameter := range r.Request.Body.Parameter {
//				keys := tools.FindAllDestStr(parameter.Key, "{{(.*?)}}")
//				if keys != nil && len(keys) > 1 {
//					for _, key := range keys {
//						if _, ok := r.Parameters.Load(key[1]); !ok {
//							r.Parameters.Store(key[1], key[0])
//						}
//					}
//				}
//				if parameter.Value == nil {
//					continue
//				}
//				values := tools.FindAllDestStr(parameter.Value.(string), "{{(.*?)}}")
//				if values != nil {
//					for _, value := range values {
//						if _, ok := r.Parameters.Load(value[1]); !ok {
//							r.Parameters.Store(value[1], value[0])
//						}
//					}
//				}
//
//			}
//		case UrlencodeMode:
//			if r.Request.Body.Parameter == nil {
//				return
//			}
//			for _, parameter := range r.Request.Body.Parameter {
//				keys := tools.FindAllDestStr(parameter.Key, "{{(.*?)}}")
//				if keys != nil {
//					for _, key := range keys {
//						if _, ok := r.Parameters.Load(key[1]); !ok {
//							r.Parameters.Store(key[1], key[0])
//						}
//					}
//				}
//				if parameter.Value == nil {
//					continue
//				}
//				values := tools.FindAllDestStr(parameter.Value.(string), "{{(.*?)}}")
//				if values != nil {
//					for _, value := range values {
//						if _, ok := r.Parameters.Load(value[1]); !ok {
//							r.Parameters.Store(value[1], value[0])
//						}
//					}
//				}
//			}
//		default:
//			if r.Request.Body.Raw == "" {
//				return
//			}
//			bodys := tools.FindAllDestStr(r.Request.Body.Raw, "{{(.*?)}}")
//			if bodys != nil {
//				for _, body := range bodys {
//					if len(body) > 1 {
//						if _, ok := r.Parameters.Load(body[1]); !ok {
//							r.Parameters.Store(body[1], body[0])
//						}
//					}
//				}
//			}
//		}
//	}
//
//}

// 将Header中的变量，都存储到接口变量中
//func (r *Api) findHeaderParameters() {
//
//	if r.Request.Header.Parameter == nil {
//		return
//	}
//	for _, varForm := range r.Request.Header.Parameter {
//		nameParameters := tools.FindAllDestStr(varForm.Key, "{{(.*?)}}")
//		for _, name := range nameParameters {
//			if _, ok := r.Parameters.Load(name[1]); !ok {
//				r.Parameters.Store(name[1], name[0])
//			}
//		}
//		if varForm.Value == nil {
//			continue
//		}
//		valueParameters := tools.FindAllDestStr(varForm.Value.(string), "{{(.*?)}}")
//		for _, value := range valueParameters {
//			if len(value) > 1 {
//				if _, ok := r.Parameters.Load(value[1]); !ok {
//					r.Parameters.Store(value[1], value[0])
//				}
//			}
//
//		}
//	}
//
//}

//func (r *Api) findAuthParameters() {
//	if r.Request.Auth != nil {
//		switch r.Request.Auth.Type {
//		case Kv:
//			if r.Request.Auth.KV.Key == "" {
//				return
//			}
//			keys := tools.FindAllDestStr(r.Request.Auth.KV.Key, "{{(.*?)}}")
//			for _, key := range keys {
//				if _, ok := r.Parameters.Load(key[1]); !ok {
//					r.Parameters.Store(key[1], key[0])
//				}
//			}
//
//			if r.Request.Auth.KV.Value == nil {
//				return
//
//			}
//
//			values := tools.FindAllDestStr(r.Request.Auth.KV.Value.(string), "{{(.*?)}}")
//			for _, value := range values {
//				if _, ok := r.Parameters.Load(value[1]); !ok {
//					r.Parameters.Store(value[1], value[0])
//				}
//			}
//		case BEarer:
//			if r.Request.Auth.Bearer.Key == "" {
//				return
//			}
//			keys := tools.FindAllDestStr(r.Request.Auth.Bearer.Key, "{{(.*?)}}")
//			for _, key := range keys {
//				if _, ok := r.Parameters.Load(key[1]); !ok {
//					r.Parameters.Store(key[1], key[0])
//				}
//			}
//		case BAsic:
//			if r.Request.Auth.Basic.UserName == "" {
//				return
//			}
//			names := tools.FindAllDestStr(r.Request.Auth.Basic.UserName, "{{(.*?)}}")
//			for _, name := range names {
//				if _, ok := r.Parameters.Load(name[1]); !ok {
//					r.Parameters.Store(name[1], name[0])
//				}
//			}
//			if r.Request.Auth.Basic.UserName == "" {
//				return
//			}
//			pws := tools.FindAllDestStr(r.Request.Auth.Basic.Password, "{{(.*?)}}")
//			for _, pw := range pws {
//				if _, ok := r.Parameters.Load(pw[1]); !ok {
//					r.Parameters.Store(pw[1], pw[0])
//				}
//			}
//		}
//	}
//}
