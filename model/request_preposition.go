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
	debugMsg := new(DebugMsg)
	debugMsg.EventId = p.Event.Id
	debugMsg.CaseId = p.Event.CaseId
	debugMsg.ParentId = p.Event.ParentId
	debugMsg.PlanId = p.Event.PlanId
	debugMsg.ReportId = p.Event.ReportId
	debugMsg.NextList = p.Event.NextList
	debugMsg.SceneId = scene.SceneId
	debugMsg.UUID = scene.Uuid.String()
	debugMsg.RequestType = p.Type

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

		a.Debug = constant.All
		debugMsg.ApiId = a.TargetId
		debugMsg.ApiName = a.Name
		debugMsg.TeamId = a.TeamId

		a.SQL.Send(a.Debug, debugMsg, mongoCollection, variable)

	}
}
