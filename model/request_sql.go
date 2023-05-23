package model

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"sync"
)

type SQL struct {
	TargetId          string               `json:"target_id"`
	Uuid              uuid.UUID            `json:"uuid"`
	Name              string               `json:"name"`
	TeamId            string               `json:"team_id"`
	TargetType        string               `json:"target_type"` // api/webSocket/tcp/grpc
	SqlString         string               `json:"sql_string"`
	MysqlDatabaseInfo MysqlDatabaseInfo    `json:"mysql_database_info"`
	Assert            []*AssertionText     `json:"assert"`  // 验证的方法(断言)
	Timeout           int64                `json:"timeout"` // 请求超时时间
	Regex             []*RegularExpression `json:"regex"`   // 正则表达式
	Debug             string               `json:"debug"`   // 是否开启Debug模式
	Configuration     *Configuration       `json:"configuration"`
	SqlVariable       *GlobalVariable      `json:"sql_variable"`    // 全局变量
	GlobalVariable    *GlobalVariable      `json:"global_variable"` // 全局变量
}

type MysqlDatabaseInfo struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Port     int32  `json:"port"`
	DbName   string `json:"db_name"`
	Charset  string `json:"charset"`
}

func (sql *SQL) Asser(results map[string]interface{}) (assertionList []AssertionMsg) {
	if sql.Assert == nil || len(sql.Assert) < 1 {
		return
	}
	for _, assert := range sql.Assert {
		if assert.IsChecked != Open {
			continue
		}
		assertionMsg := AssertionMsg{}
		if results == nil || len(results) < 1 {
			assertionMsg.Code = 10001
			assertionMsg.IsSucceed = false
			assertionMsg.Msg = fmt.Sprintf("%s不存在，断言失败", assert.Var)
			assertionList = append(assertionList, assertionMsg)
			continue
		}
		switch assert.Compare {
		case Equal:
			if value, ok := results[assert.Var]; !ok {
				assertionMsg.Code = 10001
				assertionMsg.IsSucceed = false
				assertionMsg.Msg = fmt.Sprintf("%s不存在，断言失败", assert.Var)
				assertionList = append(assertionList, assertionMsg)
				continue
			} else {
				if assert.Index == -1 {
					if value == assert.Val {
						assertionMsg.Code = 10000
						assertionMsg.IsSucceed = true
						assertionMsg.Msg = fmt.Sprintf("%s 的值等于%s, 断言成功！", assert.Var, assert.Val)
						assertionList = append(assertionList, assertionMsg)
						continue
					} else {
						assertionMsg.Code = 10001
						assertionMsg.IsSucceed = false
						assertionMsg.Msg = fmt.Sprintf("%s 的值不等于%s, 断言失败！", assert.Var, assert.Val)
						assertionList = append(assertionList, assertionMsg)
						continue
					}
				} else if len(value.([]interface{})) > assert.Index {
					if value.([]interface{})[assert.Index] == assert.Val {
						assertionMsg.Code = 10000
						assertionMsg.IsSucceed = true
						assertionMsg.Msg = fmt.Sprintf("%s 下标为%d的值等于%s, 断言成功！", assert.Var, assert.Index, assert.Val)
						assertionList = append(assertionList, assertionMsg)
						continue
					} else {
						assertionMsg.Code = 10001
						assertionMsg.IsSucceed = false
						assertionMsg.Msg = fmt.Sprintf("%s 下标为%d的值不等于%s, 断言失败！", assert.Var, assert.Index, assert.Val)
						assertionList = append(assertionList, assertionMsg)
						continue
					}
				} else {
					assertionMsg.Code = 10001
					assertionMsg.IsSucceed = false
					assertionMsg.Msg = fmt.Sprintf("%s 下标为%d的值不存在, 断言失败！", assert.Var, assert.Index)
					assertionList = append(assertionList, assertionMsg)
					continue
				}
			}
		default:
			assertionMsg.Code = 10001
			assertionMsg.IsSucceed = false
			assertionMsg.Msg = fmt.Sprintf("条件表达式%s 不存在，断言失败", assert.Compare)
			assertionList = append(assertionList, assertionMsg)
			continue

		}

	}
	return
}

func (sql *SQL) RegexSql(results map[string]interface{}, globalVar *sync.Map) (regexs []map[string]interface{}) {
	if sql.Regex == nil || len(sql.Regex) <= 0 || results == nil {
		return
	}
	for _, regex := range sql.Regex {
		if regex.IsChecked != Open {
			continue
		}
		reg := make(map[string]interface{})
		if value, ok := results[regex.Var]; ok {

			switch regex.Index {
			case -1:
				globalVar.Store(regex.Var, value)
				reg[regex.Var] = value
			default:
				if len(value.([]interface{})) > regex.Index {
					globalVar.Store(regex.Var, value.([]interface{})[regex.Index])
					reg[regex.Var] = value.([]interface{})[regex.Index]
				} else {
					globalVar.Store(regex.Var, nil)
					reg[regex.Var] = nil
				}
			}
		} else {
			globalVar.Store(regex.Var, nil)
			reg[regex.Var] = nil
		}
		regexs = append(regexs, reg)
	}
	return
}
