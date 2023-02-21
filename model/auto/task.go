package auto

// ConfigTask 任务配置
type ConfigTask struct {
	TaskType     int64  `json:"task_type" bson:"task_type"`           // 任务类型：0. 普通任务； 1. 定时任务； 2. cicd任务
	TaskMode     int64  `json:"task_mode" bson:"task_mode"`           // 1. 按用例执行
	SceneRunMode int64  `json:"scene_run_mode" bson:"scene_run_mode"` // 2. 同时执行； 1. 顺序执行
	CaseRunMode  int64  `json:"case_run_mode" bson:"case_run_mode"`   // 2. 同时执行； 1. 顺序执行
	Remark       string `json:"remark" bson:"remark"`                 // 备注
}
