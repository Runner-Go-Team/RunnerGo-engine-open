package model

import (
	"github.com/Runner-Go-Team/RunnerGo-engine-open/constant"
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
	Prepositions            []*Preposition  `json:"prepositions"`              // 前置条件
	NodesRound              [][]Event       `json:"nodes_round"`               // 事件二元数组
	ConfigTask              *ConfigTask     `json:"config_task"`               // 任务配置
	Configuration           *Configuration  `json:"configuration"`             // 场景配置
	GlobalVariable          *GlobalVariable `json:"global_variable"`           // 全局变量
	Cases                   []Scene         `json:"cases"`                     // 用例集
}

type Configuration struct {
	ParameterizedFile *ParameterizedFile `json:"parameterizedFile"`
	SceneVariable     *GlobalVariable    `json:"scene_variable"`
	Mu                sync.Mutex         `json:"mu"`
}

// VarToSceneKV 使用数据
func (c *Configuration) VarToSceneKV() []*KV {
	if c.ParameterizedFile.VariableNames.VarMapLists == nil {
		return nil
	}
	var kvList []*KV
	for k, v := range c.ParameterizedFile.VariableNames.VarMapLists {
		if v.Index >= len(v.Value) {
			v.Index = 0
		}
		var kv = new(KV)
		kv.Key = k
		if len(v.Value) > 0 {
			kv.Value = v.Value[v.Index]
		} else {
			kv.Value = ""
		}

		kvList = append(kvList, kv)
		v.Index++
	}
	return kvList
}

// SupToSub 将上一级的global添加到下一级的global中
func (g *GlobalVariable) SupToSub(variable *GlobalVariable) {
	if g.Header != nil && g.Header.Parameter != nil && len(g.Header.Parameter) > 0 {
		if variable.Header == nil {
			variable.Header = new(Header)
		}
		if variable.Header.Parameter == nil {
			variable.Header.Parameter = []*VarForm{}
		}
		for _, parameter := range g.Header.Parameter {
			if parameter.IsChecked != constant.Open {
				continue
			}
			var isExist bool
			for _, header := range variable.Header.Parameter {
				if header.IsChecked == constant.Open && parameter.Key == header.Key && header.Value == parameter.Value {
					isExist = true
				}
			}

			if !isExist {
				variable.Header.Parameter = append(variable.Header.Parameter, parameter)
			}
		}
	}

	if g.Cookie != nil && g.Cookie.Parameter != nil && len(g.Cookie.Parameter) > 0 {
		if variable.Cookie == nil {
			variable.Cookie = new(Cookie)
		}
		if variable.Cookie.Parameter == nil {
			variable.Cookie.Parameter = []*VarForm{}
		}
		for _, parameter := range g.Cookie.Parameter {
			if parameter.IsChecked != constant.Open {
				continue
			}
			var isExist bool
			for _, cookie := range variable.Cookie.Parameter {
				if cookie.IsChecked == constant.Open && parameter.Key == cookie.Key && parameter.Value == cookie.Value {
					isExist = true
				}
			}
			if !isExist {
				variable.Cookie.Parameter = append(variable.Cookie.Parameter, parameter)
			}
		}
	}

	if g.Variable != nil && len(g.Variable) > 0 {
		if variable.Variable == nil {
			variable.Variable = []*VarForm{}
		}
		for _, parameter := range g.Variable {
			if parameter.IsChecked != constant.Open {
				continue
			}
			var isExist bool
			for _, v := range variable.Variable {
				if v.IsChecked == constant.Open && v.Key == parameter.Key {
					isExist = true
				}
			}
			if !isExist {
				variable.Variable = append(variable.Variable, parameter)
			}
		}
	}

	if g.Assert != nil && len(g.Assert) > 0 {
		if variable.Assert == nil {
			variable.Assert = []*AssertionText{}
		}
		for _, parameter := range g.Assert {
			if parameter.IsChecked != constant.Open {
				continue
			}
			var isExist bool
			for _, a := range variable.Assert {
				if a.IsChecked == constant.Open && a.Var == parameter.Var && a.Val == parameter.Val && a.Compare == parameter.Compare {
					isExist = true
				}
			}
			if !isExist {
				variable.Assert = append(variable.Assert, parameter)
			}
		}
	}

}

// InitReplace 将公共变量/cookie/header/assert中的公共函数的值初始化
func (g *GlobalVariable) InitReplace() {
	if g.Variable != nil && len(g.Variable) > 0 {
		for _, kv := range g.Variable {
			if kv.IsChecked != constant.Open {
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

	if g.Header != nil && g.Header.Parameter != nil && len(g.Header.Parameter) > 0 {
		for _, parameter := range g.Header.Parameter {
			if parameter.IsChecked != constant.Open {
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
						continue
					}
					if g.Variable != nil {
						for _, variable := range g.Variable {
							if variable.IsChecked != constant.Open {
								continue
							}
							if v[1] == variable.Key {
								parameter.Value = strings.Replace(parameter.Value.(string), v[0], variable.Value.(string), -1)
							}
						}
					}
				}
			}
		}
	}
	if g.Cookie != nil && g.Cookie.Parameter != nil && len(g.Cookie.Parameter) > 0 {
		for _, parameter := range g.Cookie.Parameter {
			if parameter.IsChecked != constant.Open {
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
						continue
					}
					if g.Variable != nil {
						for _, variable := range g.Variable {
							if variable.IsChecked != constant.Open {
								continue
							}

							if v[1] == variable.Key {
								parameter.Value = strings.Replace(parameter.Value.(string), v[0], variable.Value.(string), -1)
							}
						}
					}
				}
			}
		}
	}

	if g.Assert != nil && len(g.Assert) > 0 {
		for _, asser := range g.Assert {
			if asser.IsChecked != constant.Open {
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
					continue
				}
				if g.Variable != nil {
					for _, variable := range g.Variable {
						if variable.IsChecked != constant.Open {
							continue
						}
						if v[1] == variable.Key {
							asser.Val = strings.Replace(asser.Val, v[0], variable.Value.(string), -1)
						}
					}
				}
			}
		}
	}
}
