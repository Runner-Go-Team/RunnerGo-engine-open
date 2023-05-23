package golink

import (
	"encoding/json"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/server/client"
	"go.mongodb.org/mongo-driver/mongo"
	"sync"
	"time"
)

func SqlSend(sql model.SQL, sqlInfo model.MysqlDatabaseInfo, mongoCollection *mongo.Collection, globalVar *sync.Map) (isSucceed bool, requestTime uint64, startTime, endTime time.Time) {
	var (
	//errCode         = model.NoError
	//receivedBytes   = float64(0)
	//errMsg          = ""
	//assertNum       = 0
	//assertFailedNum = 0
	)
	isSucceed = true
	db, result, err, startTime, endTime, requestTime := client.SqlRequest(sqlInfo, sql.SqlString)
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
		by, _ := json.Marshal(sqlInfo)
		if by != nil {
			results["database"] = string(by)
		}
		results["sql"] = sql.SqlString
		results["regex"] = regex
	}
	model.Insert(mongoCollection, results, middlewares.LocalIp)
	return
}
