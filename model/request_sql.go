package model

import (
	sql_client "database/sql"
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/constant"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	_ "github.com/go-sql-driver/mysql" //mysql驱动
	_ "github.com/lib/pq"              //postgres驱动
	"go.mongodb.org/mongo-driver/mongo"
	//_ "github.com/mattn/go-oci8"       //oracle驱动
	"strings"
	"sync"
	"time"
)

type SQLDetail struct {
	SqlString       string          `json:"sql_string"`
	SqlDatabaseInfo SqlDatabaseInfo `json:"sql_database_info"`
	Assert          []*SqlAssert    `json:"assert"`  // 验证的方法(断言)
	Timeout         int64           `json:"timeout"` // 请求超时时间
	Regex           []*SqlRegex     `json:"regex"`   // 关联提取
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

func (sql *SQLDetail) Send(debug string, debugMsg map[string]interface{}, mongoCollection *mongo.Collection, globalVar *sync.Map) (isSucceed bool, requestTime uint64, startTime, endTime time.Time) {
	isSucceed = true
	sql.SqlString = strings.ToLower(strings.TrimSpace(strings.NewReplacer("\r", " ", "\n", " ").Replace(sql.SqlString)))
	db, result, err, startTime, endTime, requestTime := sql.Request()
	responseTime := endTime.Format("2006-01-02 15:04:05")
	defer func() {
		if db != nil {
			db.Close()
		}
	}()
	if err != nil {
		isSucceed = false
	}

	assertionList := sql.Asser(result)
	for _, assert := range assertionList {
		if assert.IsSucceed == false {
			isSucceed = false
		}
	}
	regex := sql.RegexSql(result, globalVar)
	switch debug {
	case constant.All:
		if err != nil {
			debugMsg["response_body"] = err.Error()
		} else {
			debugMsg["response_body"] = result
		}
		debugMsg["request_time"] = requestTime / uint64(time.Millisecond)

		debugMsg["assert"] = assertionList
		debugMsg["status"] = constant.Success
		by, _ := json.Marshal(sql.SqlDatabaseInfo)
		if by != nil {
			debugMsg["request_url"] = string(by)
		}
		debugMsg["request_body"] = sql.SqlString
		debugMsg["regex"] = regex
		debugMsg["response_time"] = responseTime
		Insert(mongoCollection, debugMsg, middlewares.LocalIp)
	case constant.OnlyError:
		if isSucceed {
			return
		}
		debugMsg["response_body"] = err.Error()
		debugMsg["request_time"] = requestTime / uint64(time.Millisecond)

		debugMsg["assert"] = assertionList
		debugMsg["status"] = constant.Success
		by, _ := json.Marshal(sql.SqlDatabaseInfo)
		if by != nil {
			debugMsg["request_url"] = string(by)
		}
		debugMsg["request_body"] = sql.SqlString
		debugMsg["regex"] = regex
		debugMsg["response_time"] = responseTime
		Insert(mongoCollection, debugMsg, middlewares.LocalIp)
	case constant.OnlySuccess:
		if !isSucceed {
			return
		}
		debugMsg["response_body"] = result
		debugMsg["request_time"] = requestTime / uint64(time.Millisecond)

		debugMsg["assert"] = assertionList
		debugMsg["status"] = constant.Success
		by, _ := json.Marshal(sql.SqlDatabaseInfo)
		if by != nil {
			debugMsg["request_url"] = string(by)
		}
		debugMsg["request_body"] = sql.SqlString
		debugMsg["regex"] = regex
		debugMsg["response_time"] = responseTime
		Insert(mongoCollection, debugMsg, middlewares.LocalIp)
	}
	return
}

func (sql *SQLDetail) Request() (db *sql_client.DB, result map[string]interface{}, err error, startTime, endTime time.Time, requestTime uint64) {
	db, err = sql.init()
	if db == nil || err != nil {
		return
	}
	//strings.EqualFold
	startTime = time.Now()
	result = make(map[string]interface{})
	if strings.HasPrefix(sql.SqlString, "select") {
		rows, errQuery := db.Query(sql.SqlString)
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
		results, errExec := db.Exec(sql.SqlString)
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

func (sql *SQLDetail) init() (db *sql_client.DB, err error) {
	var dsn string
	sqlInfo := sql.SqlDatabaseInfo
	switch sqlInfo.Type {
	case "oracle":
		sqlInfo.Type = "oci8"
		dsn = fmt.Sprintf("%s:%s@%s:%d/%s?%s", sqlInfo.User, sqlInfo.Password, sqlInfo.Host, sqlInfo.Port, sqlInfo.DbName, sqlInfo.Charset)
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", sqlInfo.User, sqlInfo.Password, sqlInfo.Host, sqlInfo.Port, sqlInfo.DbName, sqlInfo.Charset)
	case "postgresql":
		sqlInfo.Type = "postgres"
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", sqlInfo.Host, sqlInfo.Port, sqlInfo.User, sqlInfo.Password, sqlInfo.DbName)
	}

	db, err = sql_client.Open(sqlInfo.Type, dsn)
	return
}

func (sql *SQLDetail) TestConnection() (db *sql_client.DB, err error) {
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

func (sql *SQLDetail) Asser(results map[string]interface{}) (assertionList []AssertionMsg) {
	if sql.Assert == nil || len(sql.Assert) < 1 {
		return
	}
	for _, assert := range sql.Assert {
		if assert.IsChecked != constant.Open {
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
		case constant.Equal:

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

func (sql *SQLDetail) RegexSql(results map[string]interface{}, globalVar *sync.Map) (regexs []map[string]interface{}) {
	if sql.Regex == nil || len(sql.Regex) <= 0 || results == nil {
		return
	}
	for _, regex := range sql.Regex {
		if regex.IsChecked != constant.Open {
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
