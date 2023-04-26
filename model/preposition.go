package model

import (
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model/js"
	"sync"
)

type Preposition struct {
	Type      string `json:"type"`
	ValueType string `json:"value_type"`
	Key       string `json:"key"`
	Scope     int32  `json:"scope"`
	JsScript  string `json:"js_script"`
	Event     Event  `json:"event"`
}

func (p *Preposition) Exec(globalVariable *sync.Map, tempVariable *GlobalVariable) {
	if p == nil {
		return
	}
	var val interface{}
	switch p.Type {
	case JSMode:
		if p.JsScript == NILSTRING {
			return
		}
		value := js.RunJs(p.JsScript)
		switch p.ValueType {
		case StringType:
			val = value.String()
		case BooleanType:
			val = value.Boolean()
		case IntegerType:
			val = value.BigInt()
		case FloatType:
			val = value.IsFloat64Array()
		}
		switch p.Scope {
		case 1:
			globalVariable.Store(p.Key, value)
		case 2:
			varForm := &VarForm{
				IsChecked: Open,
				Key:       p.Key,
				Value:     val,
			}
			tempVariable.Variable = append(tempVariable.Variable, varForm)

		}
	}
}
