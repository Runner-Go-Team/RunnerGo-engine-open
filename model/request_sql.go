package model

import (
	sql_client "database/sql"
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"strings"
	"sync"
	"time"
)

type SQL struct {
	TargetId        string          `json:"target_id"`
	Uuid            uuid.UUID       `json:"uuid"`
	Name            string          `json:"name"`
	TeamId          string          `json:"team_id"`
	TargetType      string          `json:"target_type"` // api/webSocket/tcp/grpc
	SqlString       string          `json:"sql_string"`
	SqlDatabaseInfo SqlDatabaseInfo `json:"sql_database_info"`
	Assert          []*SqlAssert    `json:"assert"`  // 验证的方法(断言)
	Timeout         int64           `json:"timeout"` // 请求超时时间
	Regex           []*SqlRegex     `json:"regex"`   // 关联提取
	Debug           string          `json:"debug"`   // 是否开启Debug模式
	Configuration   *Configuration  `json:"configuration"`
	SqlVariable     *GlobalVariable `json:"sql_variable"`    // 全局变量
	GlobalVariable  *GlobalVariable `json:"global_variable"` // 全局变量
}
type SqlDatabaseInfo struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Port     int32  `json:"port"`
	DbName   string `json:"db_name"`
	Charset  string `json:"charset"`
}
type SqlAssert struct {
	IsChecked int    `json:"is_checked"`
	Field     string `json:"field"`
	Compare   string `json:"compare"`
	Val       string `json:"val"`
	Index     int    `json:"index"` // 断言时提取第几个值
}

type SqlRegex struct {
	IsChecked int    `json:"is_checked"` // 1 选中, -1未选
	Var       string `json:"var"`
	Field     string `json:"field"`
	Index     int    `json:"index"` // 正则时提取第几个值
}

func (sql *SQL) Send(mongoCollection *mongo.Collection, globalVar *sync.Map) (isSucceed bool, requestTime uint64, startTime, endTime time.Time) {
	isSucceed = true
	db, result, err, startTime, endTime, requestTime := sql.Request()
	defer db.Close()
	if err != nil {
		isSucceed = false
	}
	results := make(map[string]interface{})
	assertionList := sql.Asser(result)
	for _, assert := range assertionList {
		if assert.IsSucceed == false {
			isSucceed = false
		}
	}
	regex := sql.RegexSql(result, globalVar)
	if sql.Debug == "all" {
		results["team_id"] = sql.TeamId
		results["sql_name"] = sql.Name
		results["target_id"] = sql.TargetId
		results["uuid"] = sql.Uuid.String()
		if err != nil {
			results["err"] = err.Error()
		} else {
			results["err"] = ""
		}

		results["request_time"] = requestTime / uint64(time.Millisecond)
		results["sql_result"] = result
		results["assertion"] = assertionList
		results["status"] = isSucceed
		by, _ := json.Marshal(sql.SqlDatabaseInfo)
		if by != nil {
			results["database"] = string(by)
		}
		results["sql"] = sql.SqlString
		results["regex"] = regex
	}
	Insert(mongoCollection, results, middlewares.LocalIp)
	return
}

func (sql *SQL) Request() (db *sql_client.DB, result map[string]interface{}, err error, startTime, endTime time.Time, requestTime uint64) {
	db, err = sql.init()
	if db == nil || err != nil {
		return
	}
	// 不区分大小写,去除空格，去除回车符/换行符
	sqls := strings.ToLower(strings.TrimSpace(strings.NewReplacer("\r", "", "\n", "").Replace(sql.SqlString)))
	//strings.EqualFold
	startTime = time.Now()
	result = make(map[string]interface{})
	if strings.HasPrefix(sqls, "select") {
		rows, errQuery := db.Query(sqls)
		requestTime = uint64(time.Since(startTime))
		if errQuery != nil || rows == nil {
			err = errQuery
			return
		}
		defer rows.Close()
		cols, _ := rows.Columns()
		values := make([]sql_client.RawBytes, len(cols))
		scans := make([]interface{}, len(cols))
		if cols != nil {
			for i := range cols {
				scans[i] = &values[i]
			}
		}
		resultMap := make(map[string][]string)
		for rows.Next() {
			if err := rows.Scan(scans...); err != nil {
				continue
			}
			for j, v := range values {
				resultMap[cols[j]] = append(resultMap[cols[j]], string(v))
			}
		}
		for k, v := range resultMap {
			result[k] = v
		}
		return

	} else {
		results, errExec := db.Exec(sqls)
		requestTime = uint64(time.Since(startTime))
		if errExec != nil || result == nil {
			err = errExec
			return
		}

		row, errExec := results.RowsAffected()
		if errExec != nil {
			return
		}
		result["rows_affected"] = row
		last, errExec := results.LastInsertId()
		if errExec != nil {
			return
		}
		result["last_insert_id"] = last
		return
	}
}

func (sql *SQL) init() (db *sql_client.DB, err error) {
	var dsn string
	sqlInfo := sql.SqlDatabaseInfo
	switch sqlInfo.Type {
	case "oracle":
		sqlInfo.Type = "oci8"
		dsn = fmt.Sprintf("%s:%s@%s:%d/%s?%s", sqlInfo.User, sqlInfo.Password, sqlInfo.Host, sqlInfo.Port, sqlInfo.DbName, sqlInfo.Charset)
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", sqlInfo.User, sqlInfo.Password, sqlInfo.Host, sqlInfo.Port, sqlInfo.DbName, sqlInfo.Charset)
	case "postgres":
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=verify-full", sqlInfo.User, sqlInfo.Password, sqlInfo.Host, sqlInfo.Port, sqlInfo.DbName)
	}

	db, err = sql_client.Open(sqlInfo.Type, dsn)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("%s数据库连接失败： %s", sqlInfo.Type, err.Error()))
		return
	}
	return
}

func (sql *SQL) TestConnection() (db *sql_client.DB, err error) {
	db, err = sql.init()
	if err != nil {
		return
	}
	err = db.Ping()
	if err != nil {
		return
	}
	return
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
