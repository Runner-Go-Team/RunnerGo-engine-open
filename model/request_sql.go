package model

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"sync"
)

type SQL struct {
	TargetId          string            `json:"target_id"`
	Uuid              uuid.UUID         `json:"uuid"`
	Name              string            `json:"name"`
	TeamId            string            `json:"team_id"`
	TargetType        string            `json:"target_type"` // api/webSocket/tcp/grpc
	SqlString         string            `json:"sql_string"`
	MysqlDatabaseInfo MysqlDatabaseInfo `json:"mysql_database_info"`
	Assert            []*MysqlAssert    `json:"assert"`  // 验证的方法(断言)
	Timeout           int64             `json:"timeout"` // 请求超时时间
	Regex             []*MysqlRegex     `json:"regex"`   // 关联提取
	Debug             string            `json:"debug"`   // 是否开启Debug模式
	Configuration     *Configuration    `json:"configuration"`
	SqlVariable       *GlobalVariable   `json:"sql_variable"`    // 全局变量
	GlobalVariable    *GlobalVariable   `json:"global_variable"` // 全局变量
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
type MysqlAssert struct {
	IsChecked int    `json:"is_checked"`
	Field     string `json:"field"`
	Compare   string `json:"compare"`
	Val       string `json:"val"`
	Index     int    `json:"index"` // 断言时提取第几个值
}

type MysqlRegex struct {
	IsChecked int    `json:"is_checked"` // 1 选中, -1未选
	Var       string `json:"var"`
	Field     string `json:"field"`
	Index     int    `json:"index"` // 正则时提取第几个值
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
			assertionMsg.Msg = fmt.Sprintf("%s不存在，断言失败", assert.Field)
			assertionList = append(assertionList, assertionMsg)
			continue
		}
		switch assert.Compare {
		case Equal:

			if value, ok := results[assert.Field]; !ok {
				assertionMsg.Code = 10001
				assertionMsg.IsSucceed = false
				assertionMsg.Msg = fmt.Sprintf("%s不存在，断言失败", assert.Field)
				assertionList = append(assertionList, assertionMsg)
				continue
			} else {
				switch fmt.Sprintf("%T", value) {
				case "[]string":
					if assert.Index == -1 {
						if value == assert.Val {
							assertionMsg.Code = 10000
							assertionMsg.IsSucceed = true
							assertionMsg.Msg = fmt.Sprintf("%s 的值等于%s, 断言成功！", assert.Field, assert.Val)
							assertionList = append(assertionList, assertionMsg)
							continue
						} else {
							assertionMsg.Code = 10001
							assertionMsg.IsSucceed = false
							assertionMsg.Msg = fmt.Sprintf("%s 的值不等于%s, 断言失败！", assert.Field, assert.Val)
							assertionList = append(assertionList, assertionMsg)
							continue
						}
					} else if len(value.([]string)) > assert.Index {
						if value.([]string)[assert.Index] == assert.Val {
							assertionMsg.Code = 10000
							assertionMsg.IsSucceed = true
							assertionMsg.Msg = fmt.Sprintf("%s 下标为%d的值等于%s, 断言成功！", assert.Field, assert.Index, assert.Val)
							assertionList = append(assertionList, assertionMsg)
							continue
						} else {
							assertionMsg.Code = 10001
							assertionMsg.IsSucceed = false
							assertionMsg.Msg = fmt.Sprintf("%s 下标为%d的值不等于%s, 断言失败！", assert.Field, assert.Index, assert.Val)
							assertionList = append(assertionList, assertionMsg)
							continue
						}
					} else {
						assertionMsg.Code = 10001
						assertionMsg.IsSucceed = false
						assertionMsg.Msg = fmt.Sprintf("%s 下标为%d的值不存在, 断言失败！", assert.Field, assert.Index)
						assertionList = append(assertionList, assertionMsg)
						continue
					}
				case "[]int":
					if assert.Index == -1 {
						if value == assert.Val {
							assertionMsg.Code = 10000
							assertionMsg.IsSucceed = true
							assertionMsg.Msg = fmt.Sprintf("%s 的值等于%s, 断言成功！", assert.Field, assert.Val)
							assertionList = append(assertionList, assertionMsg)
							continue
						} else {
							assertionMsg.Code = 10001
							assertionMsg.IsSucceed = false
							assertionMsg.Msg = fmt.Sprintf("%s 的值不等于%s, 断言失败！", assert.Field, assert.Val)
							assertionList = append(assertionList, assertionMsg)
							continue
						}
					} else if len(value.([]int)) > assert.Index {
						if fmt.Sprintf("%d", value.([]int)[assert.Index]) == assert.Val {
							assertionMsg.Code = 10000
							assertionMsg.IsSucceed = true
							assertionMsg.Msg = fmt.Sprintf("%s 下标为%d的值等于%s, 断言成功！", assert.Field, assert.Index, assert.Val)
							assertionList = append(assertionList, assertionMsg)
							continue
						} else {
							assertionMsg.Code = 10001
							assertionMsg.IsSucceed = false
							assertionMsg.Msg = fmt.Sprintf("%s 下标为%d的值不等于%s, 断言失败！", assert.Field, assert.Index, assert.Val)
							assertionList = append(assertionList, assertionMsg)
							continue
						}
					} else {
						assertionMsg.Code = 10001
						assertionMsg.IsSucceed = false
						assertionMsg.Msg = fmt.Sprintf("%s 下标为%d的值不存在, 断言失败！", assert.Field, assert.Index)
						assertionList = append(assertionList, assertionMsg)
						continue
					}
				case "[]float64":
					if assert.Index == -1 {
						if value == assert.Val {
							assertionMsg.Code = 10000
							assertionMsg.IsSucceed = true
							assertionMsg.Msg = fmt.Sprintf("%s 的值等于%s, 断言成功！", assert.Field, assert.Val)
							assertionList = append(assertionList, assertionMsg)
							continue
						} else {
							assertionMsg.Code = 10001
							assertionMsg.IsSucceed = false
							assertionMsg.Msg = fmt.Sprintf("%s 的值不等于%s, 断言失败！", assert.Field, assert.Val)
							assertionList = append(assertionList, assertionMsg)
							continue
						}
					} else if len(value.([]float64)) > assert.Index {
						if fmt.Sprintf("%v", value.([]float64)[assert.Index]) == assert.Val {
							assertionMsg.Code = 10000
							assertionMsg.IsSucceed = true
							assertionMsg.Msg = fmt.Sprintf("%s 下标为%d的值等于%s, 断言成功！", assert.Field, assert.Index, assert.Val)
							assertionList = append(assertionList, assertionMsg)
							continue
						} else {
							assertionMsg.Code = 10001
							assertionMsg.IsSucceed = false
							assertionMsg.Msg = fmt.Sprintf("%s 下标为%d的值不等于%s, 断言失败！", assert.Field, assert.Index, assert.Val)
							assertionList = append(assertionList, assertionMsg)
							continue
						}
					} else {
						assertionMsg.Code = 10001
						assertionMsg.IsSucceed = false
						assertionMsg.Msg = fmt.Sprintf("%s 下标为%d的值不存在, 断言失败！", assert.Field, assert.Index)
						assertionList = append(assertionList, assertionMsg)
						continue
					}
				case "[]bool":
					if assert.Index == -1 {
						if value == assert.Val {
							assertionMsg.Code = 10000
							assertionMsg.IsSucceed = true
							assertionMsg.Msg = fmt.Sprintf("%s 的值等于%s, 断言成功！", assert.Field, assert.Val)
							assertionList = append(assertionList, assertionMsg)
							continue
						} else {
							assertionMsg.Code = 10001
							assertionMsg.IsSucceed = false
							assertionMsg.Msg = fmt.Sprintf("%s 的值不等于%s, 断言失败！", assert.Field, assert.Val)
							assertionList = append(assertionList, assertionMsg)
							continue
						}
					} else if len(value.([]bool)) > assert.Index {
						if fmt.Sprintf("%v", value.([]int)[assert.Index]) == assert.Val {
							assertionMsg.Code = 10000
							assertionMsg.IsSucceed = true
							assertionMsg.Msg = fmt.Sprintf("%s 下标为%d的值等于%s, 断言成功！", assert.Field, assert.Index, assert.Val)
							assertionList = append(assertionList, assertionMsg)
							continue
						} else {
							assertionMsg.Code = 10001
							assertionMsg.IsSucceed = false
							assertionMsg.Msg = fmt.Sprintf("%s 下标为%d的值不等于%s, 断言失败！", assert.Field, assert.Index, assert.Val)
							assertionList = append(assertionList, assertionMsg)
							continue
						}
					} else {
						assertionMsg.Code = 10001
						assertionMsg.IsSucceed = false
						assertionMsg.Msg = fmt.Sprintf("%s 下标为%d的值不存在, 断言失败！", assert.Field, assert.Index)
						assertionList = append(assertionList, assertionMsg)
						continue
					}

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
		if value, ok := results[regex.Field]; ok {

			switch regex.Index {
			case -1:
				globalVar.Store(regex.Var, value)
				reg[regex.Var] = value
			default:
				switch fmt.Sprintf("%T", value) {
				case "[]string":
					if len(value.([]string)) > regex.Index {
						globalVar.Store(regex.Var, value.([]string)[regex.Index])
						reg[regex.Var] = value.([]string)[regex.Index]
					} else {
						globalVar.Store(regex.Var, nil)
						reg[regex.Var] = nil
					}
				case "[]int":
					if len(value.([]int)) > regex.Index {
						globalVar.Store(regex.Var, value.([]int)[regex.Index])
						reg[regex.Var] = value.([]int)[regex.Index]
					} else {
						globalVar.Store(regex.Var, nil)
						reg[regex.Var] = nil
					}
				case "[]float64":
					if len(value.([]float64)) > regex.Index {
						globalVar.Store(regex.Var, value.([]float64)[regex.Index])
						reg[regex.Var] = value.([]float64)[regex.Index]
					} else {
						globalVar.Store(regex.Var, nil)
						reg[regex.Var] = nil
					}
				case "[]bool":
					if len(value.([]bool)) > regex.Index {
						globalVar.Store(regex.Var, value.([]bool)[regex.Index])
						reg[regex.Var] = value.([]bool)[regex.Index]
					} else {
						globalVar.Store(regex.Var, nil)
						reg[regex.Var] = nil
					}

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
