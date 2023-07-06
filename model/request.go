package model

import (
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/constant"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/tools"
	uuid "github.com/satori/go.uuid"
	"strings"
	"sync"
)

// Api 请求数据
type Api struct {
	TargetId       string          `json:"target_id"`
	Uuid           uuid.UUID       `json:"uuid"`
	Name           string          `json:"name"`
	TeamId         string          `json:"team_id"`
	TargetType     string          `json:"target_type"` // api/webSocket/tcp/grpc
	Debug          string          `json:"debug"`       // 是否开启Debug模式
	Request        RequestHttp     `json:"request"`
	SQL            SQLDetail       `json:"sql_detail"`
	TCP            TCPDetail       `json:"tcp_detail"`
	Ws             WebsocketDetail `json:"ws_detail"`
	DubboDetail    DubboDetail     `json:"dubbo_detail"`
	Configuration  *Configuration  `json:"configuration"`   // 场景设置
	GlobalVariable *GlobalVariable `json:"global_variable"` // 全局变量
	ApiVariable    *GlobalVariable `json:"api_variable"`
}

func (api *Api) GlobalToRequest() {

	if api.ApiVariable.Cookie != nil && len(api.ApiVariable.Cookie.Parameter) > 0 {
		if api.Request.Cookie == nil {
			api.Request.Cookie = new(Cookie)
		}
		if api.Request.Cookie.Parameter == nil {
			api.Request.Cookie.Parameter = []*VarForm{}
		}
		for _, parameter := range api.ApiVariable.Cookie.Parameter {
			if parameter.IsChecked != constant.Open {
				continue
			}
			var isExist bool
			for _, value := range api.Request.Cookie.Parameter {
				if value.IsChecked == constant.Open && parameter.Key == value.Key && parameter.Value == value.Value {
					isExist = true
				}
			}
			if isExist {
				continue
			}
			api.Request.Cookie.Parameter = append(api.Request.Cookie.Parameter, parameter)
		}
	}

	if api.ApiVariable.Header != nil && len(api.ApiVariable.Header.Parameter) > 0 {
		if api.Request.Header == nil {
			api.Request.Header = new(Header)
		}
		if api.Request.Header.Parameter == nil {
			api.Request.Header.Parameter = []*VarForm{}
		}
		for _, parameter := range api.ApiVariable.Header.Parameter {
			if parameter.IsChecked != constant.Open {
				continue
			}
			var isExist bool
			for _, value := range api.Request.Header.Parameter {
				if value.IsChecked == constant.Open && parameter.Key == value.Key && parameter.Value == parameter.Value {
					isExist = true
				}
			}
			if isExist {
				continue
			}
			api.Request.Header.Parameter = append(api.Request.Header.Parameter, parameter)

		}
	}

	if api.ApiVariable.Assert != nil && len(api.ApiVariable.Assert) > 0 {
		if api.Request.Assert == nil {
			api.Request.Assert = []*AssertionText{}
		}
		for _, parameter := range api.ApiVariable.Assert {
			if parameter.IsChecked != constant.Open {
				continue
			}
			var isExist bool
			for _, asser := range api.Request.Assert {
				if asser.IsChecked == constant.Open && parameter.ResponseType == asser.ResponseType && parameter.Compare == asser.Compare && parameter.Val == asser.Val && parameter.Var == asser.Var {
					isExist = true
				}
			}
			if isExist {
				continue
			}
			api.Request.Assert = append(api.Request.Assert, parameter)

		}
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
	if r.Request.Assert == nil {
		r.Request.Assert = r.Configuration.SceneVariable.Assert
		return
	}
	for _, assert := range r.Configuration.SceneVariable.Assert {
		r.Request.Assert = append(r.Request.Assert, assert)
	}
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
	case constant.NoneMode:
	case constant.FormMode:
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

	case constant.UrlencodeMode:
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
	case constant.Kv:
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

	case constant.BEarer:
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
	case constant.BAsic:
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
	if r.Request.Assert == nil || len(r.Request.Assert) <= 0 {
		return
	}
	for _, assert := range r.Request.Assert {
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
