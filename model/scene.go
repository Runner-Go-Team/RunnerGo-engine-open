package model

import (
	"github.com/Runner-Go-Team/RunnerGo-engine-open/tools"
	uuid "github.com/satori/go.uuid"
	"strings"
	"sync"
)

// Scene 场景结构体
type Scene struct {
	PlanId                  string          `json:"plan_id"`
	SceneId                 string          `json:"scene_id"`   // 场景Id
	IsChecked               int64           `json:"is_checked"` // 是否启用
	ParentId                string          `json:"parentId"`
	CaseId                  string          `json:"case_id"`
	Partition               int32           `json:"partition"`
	MachineNum              int64           `json:"machine_num"` // 使用的机器数量
	Uuid                    uuid.UUID       `json:"uuid"`
	ReportId                string          `json:"report_id"`
	TeamId                  string          `json:"team_id"`
	SceneName               string          `json:"scene_name"` // 场景名称
	Version                 int64           `json:"version"`
	Debug                   string          `json:"debug"`
	EnablePlanConfiguration bool            `json:"enable_plan_configuration"` // 是否启用计划的任务配置，默认为true，
	Nodes                   []Event         `json:"nodes"`                     // 事件列表
	NodesRound              [][]Event       `json:"nodes_round"`               // 事件二元数组
	ConfigTask              *ConfigTask     `json:"config_task"`               // 任务配置
	Configuration           *Configuration  `json:"configuration"`             // 场景配置
	GlobalVariable          *GlobalVariable `json:"global_variable"`           // 全局变量
	Cases                   []Scene         `json:"cases"`                     // 用例集
}

type Configuration struct {
	ParameterizedFile *ParameterizedFile `json:"parameterizedFile"`
	GlobalVariable    *GlobalVariable    `json:"global_variable"`
	Mu                sync.Mutex         `json:"mu" bson:"mu"`
}

// VarToSceneKV 使用数据
func (c *Configuration) VarToSceneKV() []*KV {
	if c.ParameterizedFile.VariableNames.VarMapList == nil {
		return nil
	}
	var kvList []*KV
	for k, v := range c.ParameterizedFile.VariableNames.VarMapList {
		if c.ParameterizedFile.VariableNames.Index >= len(v) {
			c.ParameterizedFile.VariableNames.Index = 0
		}
		var kv = new(KV)
		kv.Key = k
		kv.Value = v[c.ParameterizedFile.VariableNames.Index]
		kvList = append(kvList, kv)
	}
	c.ParameterizedFile.VariableNames.Index++
	return kvList
}

func (g *GlobalVariable) GlobalToLocal(variable *GlobalVariable) {
	if g.Header != nil && g.Header.Parameter != nil && len(g.Header.Parameter) > 0 {
		for _, parameter := range g.Header.Parameter {
			if parameter.IsChecked != Open {
				continue
			}
			if parameter.Value != nil {
				values := tools.FindAllDestStr(parameter.Value.(string), "{{(.*?)}}")
				if values == nil {
					continue
				}

				for _, v := range values {
					if len(v) < 2 {
						continue
					}
					realVar := tools.ParsFunc(v[1])
					if realVar != v[1] {
						parameter.Value = strings.Replace(parameter.Value.(string), v[0], realVar, -1)
					}
				}
			}
			var nonexistence bool
			for _, header := range g.Header.Parameter {
				if header.IsChecked == Open && parameter.Key == header.Key {
					nonexistence = true
				}
			}
			if !nonexistence {
				variable.Header.Parameter = append(variable.Header.Parameter, parameter)
			}

		}
	}
	if g.Cookie != nil && g.Cookie.Parameter != nil && len(g.Cookie.Parameter) > 0 {
		for _, parameter := range g.Cookie.Parameter {
			if parameter.IsChecked != Open {
				continue
			}
			if parameter.Value != nil {
				values := tools.FindAllDestStr(parameter.Value.(string), "{{(.*?)}}")
				if values == nil {
					continue
				}

				for _, v := range values {
					if len(v) < 2 {
						continue
					}
					realVar := tools.ParsFunc(v[1])
					if realVar != v[1] {
						parameter.Value = strings.Replace(parameter.Value.(string), v[0], realVar, -1)
					}
				}
			}
			var nonexistence bool
			for _, header := range variable.Cookie.Parameter {
				if header.IsChecked == Open && parameter.Key == header.Key {
					nonexistence = true
				}
			}
			if !nonexistence {
				variable.Cookie.Parameter = append(variable.Cookie.Parameter, parameter)
			}

		}
	}

	if g.Variable != nil && len(g.Variable) > 0 {
		for _, kv := range g.Variable {
			if kv.IsChecked != Open {
				continue
			}
			if kv.Value != nil {
				values := tools.FindAllDestStr(kv.Value.(string), "{{(.*?)}}")
				if values == nil {
					continue
				}

				for _, v := range values {
					if len(v) < 2 {
						continue
					}
					realVar := tools.ParsFunc(v[1])
					if realVar != v[1] {
						kv.Value = strings.Replace(kv.Value.(string), v[0], realVar, -1)
					}
				}
			}
			var nonexistence bool
			for _, varKV := range variable.Variable {
				if varKV.IsChecked == Open && varKV.Key == kv.Key {
					nonexistence = true
				}
			}
			if !nonexistence {
				variable.Variable = append(variable.Variable, kv)
			}
		}
	}

	if g.Assert != nil && len(g.Assert) > 0 {
		for _, asser := range g.Assert {
			if asser.IsChecked != Open {
				continue
			}
			var nonexistence bool
			for _, asser2 := range variable.Assert {
				if asser2.IsChecked == Open && asser.ResponseType == asser2.ResponseType && asser.Compare == asser2.Compare && asser.Var == asser2.Var && asser.Val == asser2.Val {
					nonexistence = true
				}
			}
			if !nonexistence {
				variable.Assert = append(variable.Assert, asser)
			}
		}
	}

}

func (g *GlobalVariable) InitReplace() {
	if len(g.Header.Parameter) > 0 {
		for _, parameter := range g.Header.Parameter {
			if parameter.IsChecked != Open {
				continue
			}
			if parameter.Value != nil {
				values := tools.FindAllDestStr(parameter.Value.(string), "{{(.*?)}}")
				if values == nil {
					continue
				}

				for _, v := range values {
					if len(v) < 2 {
						continue
					}
					realVar := tools.ParsFunc(v[1])
					if realVar != v[1] {
						parameter.Value = strings.Replace(parameter.Value.(string), v[0], realVar, -1)
					}
				}
			}
		}
	}
	if len(g.Cookie.Parameter) > 0 {
		for _, parameter := range g.Cookie.Parameter {
			if parameter.IsChecked != Open {
				continue
			}
			if parameter.Value != nil {
				values := tools.FindAllDestStr(parameter.Value.(string), "{{(.*?)}}")
				if values == nil {
					continue
				}

				for _, v := range values {
					if len(v) < 2 {
						continue
					}
					realVar := tools.ParsFunc(v[1])
					if realVar != v[1] {
						parameter.Value = strings.Replace(parameter.Value.(string), v[0], realVar, -1)
					}
				}
			}
		}
	}

	if len(g.Variable) > 0 {
		for _, kv := range g.Variable {
			if kv.IsChecked != Open {
				continue
			}
			if kv.Value != nil {
				values := tools.FindAllDestStr(kv.Value.(string), "{{(.*?)}}")
				if values == nil {
					continue
				}

				for _, v := range values {
					if len(v) < 2 {
						continue
					}
					realVar := tools.ParsFunc(v[1])
					if realVar != v[1] {
						kv.Value = strings.Replace(kv.Value.(string), v[0], realVar, -1)
					}
				}
			}
		}
	}
	if len(g.Assert) > 0 {
		for _, asser := range g.Assert {
			if asser.IsChecked != Open {
				continue
			}
			values := tools.FindAllDestStr(asser.Val, "{{(.*?)}}")
			if values == nil {
				continue
			}

			for _, v := range values {
				if len(v) < 2 {
					continue
				}
				realVar := tools.ParsFunc(v[1])
				if realVar != v[1] {
					asser.Val = strings.Replace(asser.Val, v[0], realVar, -1)
				}
			}
		}
	}
}

func (g *GlobalVariable) GlobalToRequest(api Api) {
	if api.TargetType != "api" {
		return
	}

	if g.Cookie != nil && g.Cookie.Parameter != nil && len(g.Cookie.Parameter) > 0 {
		for _, parameter := range g.Cookie.Parameter {
			if parameter.IsChecked != Open {
				continue
			}
			if api.Request.Cookie == nil {
				api.Request.Cookie = new(Cookie)
			}
			if api.Request.Cookie.Parameter == nil {
				api.Request.Cookie.Parameter = []*VarForm{}
			}
			var isExist bool
			for _, value := range api.Request.Cookie.Parameter {
				if value.IsChecked == Open && parameter.Key == value.Key {
					isExist = true
				}
			}
			if !isExist {
				continue
			}
			api.Request.Cookie.Parameter = append(api.Request.Cookie.Parameter, parameter)

		}
	}

	if api.GlobalVariable.Header != nil && api.GlobalVariable.Header.Parameter != nil && len(api.GlobalVariable.Header.Parameter) > 0 {
		for _, parameter := range api.GlobalVariable.Header.Parameter {
			if parameter.IsChecked != Open {
				continue
			}
			if api.Request.Header == nil {
				api.Request.Header = new(Header)
			}
			if api.Request.Header.Parameter == nil {
				api.Request.Header.Parameter = []*VarForm{}
			}
			var isExist bool
			for _, value := range api.Request.Header.Parameter {
				if value.IsChecked == Open && parameter.Key == value.Key {
					isExist = true
				}
			}
			if !isExist {
				continue
			}
			api.Request.Header.Parameter = append(api.Request.Header.Parameter, parameter)

		}
	}

	if api.GlobalVariable.Assert != nil && len(api.GlobalVariable.Assert) > 0 {
		for _, parameter := range api.GlobalVariable.Assert {
			if parameter.IsChecked != Open {
				continue
			}
			if api.Assert == nil {
				api.Request.Header = new(Header)
			}
			var isExist bool
			for _, assert := range api.Assert {
				if assert.IsChecked == Open && parameter.ResponseType == assert.ResponseType && parameter.Compare == assert.Compare && parameter.Val == assert.Val && parameter.Var == assert.Var {
					isExist = true
				}
			}
			if !isExist {
				continue
			}
			api.Assert = append(api.Assert, parameter)

		}
	}

}
