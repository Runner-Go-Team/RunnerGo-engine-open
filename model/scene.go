package model

import (
	uuid "github.com/satori/go.uuid"
	"sync"
)

// Scene 场景结构体
type Scene struct {
	PlanId                  string         `json:"plan_id"`
	SceneId                 string         `json:"scene_id"`   // 场景Id
	IsChecked               int64          `json:"is_checked"` // 是否启用
	ParentId                string         `json:"parentId"`
	CaseId                  string         `json:"case_id"`
	Partition               int32          `json:"partition"`
	MachineNum              int64          `json:"machine_num"` // 使用的机器数量
	Uuid                    uuid.UUID      `json:"uuid"`
	ReportId                string         `json:"report_id"`
	TeamId                  string         `json:"team_id"`
	SceneName               string         `json:"scene_name"` // 场景名称
	Version                 int64          `json:"version"`
	Debug                   string         `json:"debug"`
	EnablePlanConfiguration bool           `json:"enable_plan_configuration"` // 是否启用计划的任务配置，默认为true，
	Nodes                   []Event        `json:"nodes"`                     // 事件列表
	NodesRound              [][]Event      `json:"nodes_round"`               // 事件二元数组
	ConfigTask              *ConfigTask    `json:"config_task"`               // 任务配置
	Configuration           *Configuration `json:"configuration"`             // 场景配置
	Variable                []*KV          `json:"variable"`                  // 场景配置
	Cases                   []Scene        `json:"cases"`                     // 用例集
}

type Configuration struct {
	ParameterizedFile *ParameterizedFile `json:"parameterizedFile" bson:"parameterizedFile"`
	Variable          []*KV              `json:"variable" bson:"variable"`
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
