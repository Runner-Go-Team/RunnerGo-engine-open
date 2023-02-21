package model

// SceneGroup 场景组结构体，分组
type SceneGroup struct {
	GroupName string `json:"groupName"`
	GroupI    string `json:"groupId"`
	Scene     Scene  `json:"scenes"`
}
