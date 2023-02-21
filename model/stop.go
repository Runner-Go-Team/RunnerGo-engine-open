package model

type Stop struct {
	TeamId    string  `json:"team_id" bson:"team_id"`
	PlanId    string  `json:"plan_id" bson:"plan_id"`
	ReportIds []int64 `json:"report_ids" bson:"report_ids"`
}

type StopScene struct {
	TeamId  string `json:"team_id"`
	SceneId string `json:"scene_id"`
}
