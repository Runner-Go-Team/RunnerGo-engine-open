package model

// Plan 计划结构体
type Plan struct {
	PlanId         string          `json:"plan_id"`    // 计划id
	PlanName       string          `json:"plan_name"`  // 计划名称
	ReportId       string          `json:"report_id" ` // 报告名称
	Partition      int32           `json:"partition"`
	TeamId         string          `json:"team_id"`         // 团队id
	ReportName     string          `json:"report_name"`     // 报告名称
	MachineNum     int64           `json:"machine_num"`     // 使用的机器数量
	ConfigTask     *ConfigTask     `json:"config_task"`     // 任务配置
	GlobalVariable *GlobalVariable `json:"global_variable"` // 全局变量
	Scene          Scene           `json:"scene"`           // 场景
	Configuration  *Configuration  `json:"configuration"`   // 场景配置
}

type GlobalVariable struct {
	Cookie   *Cookie          `json:"cookie"`
	Header   *Header          `json:"header"`
	Variable []*VarForm       `json:"variable"`
	Assert   []*AssertionText `json:"assert"` // 验证的方法(断言)
}
