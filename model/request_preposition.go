package model

import (
	"github.com/Runner-Go-Team/RunnerGo-engine-open/constant"
	"go.mongodb.org/mongo-driver/mongo"
	"sync"
)

type Preposition struct {
	Type       string `json:"type"`
	ValueType  string `json:"value_type"`
	Key        string `json:"key"`
	Scope      int32  `json:"scope"`
	IsDisabled int    `json:"is_disabled"` // 0: 不禁用， 1: 禁用
	JsScript   string `json:"js_script"`
	Event      Event  `json:"event"`
	TempMap    sync.Map
}

func (p *Preposition) Exec(scene Scene, mongoCollection *mongo.Collection, variable *sync.Map) {
	if p == nil {
		return
	}
	//var val interface{}
	switch p.Type {
	//case JSMode:
	//	if p.JsScript == NILSTRING {
	//		return
	//	}
	//	value := js.RunJs(p.JsScript)
	//	switch p.ValueType {
	//	case StringType:
	//		val = value.String()
	//	case BooleanType:
	//		val = value.Boolean()
	//	case IntegerType:
	//		val = value.BigInt()
	//	case FloatType:
	//		val = value.IsFloat64Array()
	//	}
	//	switch p.Scope {
	//	case 1:
	//		globalVariable.Store(p.Key, value)
	//	case 2:
	//		varForm := &VarForm{
	//			IsChecked: Open,
	//			Key:       p.Key,
	//			Value:     val,
	//		}
	//		tempVariable.Variable = append(tempVariable.Variable, varForm)
	//
	//	}
	case constant.MysqlMode:
		a := p.Event.Api
		debugMsg := make(map[string]interface{})
		a.Debug = constant.All
		debugMsg["api_id"] = a.TargetId
		debugMsg["event_id"] = p.Event.Id
		debugMsg["case_id"] = p.Event.CaseId
		debugMsg["parent_id"] = p.Event.ParentId
		debugMsg["plan_id"] = p.Event.PlanId
		debugMsg["report_id"] = p.Event.ReportId
		debugMsg["pre_list"] = p.Event.PreList
		debugMsg["next_list"] = p.Event.NextList
		debugMsg["api_name"] = a.Name
		debugMsg["team_id"] = a.TeamId
		debugMsg["scene_id"] = scene.SceneId
		debugMsg["uuid"] = scene.Uuid.String()
		debugMsg["request_type"] = p.Type

		a.SQL.Send(a.Debug, debugMsg, mongoCollection, variable)

	}
}
