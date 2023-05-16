package golink

import (
	"encoding/json"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/server/client"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

func SqlSend(sql model.SQL, sqlInfo model.MysqlDatabaseInfo, mongoCollection *mongo.Collection) (isSucceed bool, requestTime uint64, startTime, endTime time.Time) {
	var (
	//errCode         = model.NoError
	//receivedBytes   = float64(0)
	//errMsg          = ""
	//assertNum       = 0
	//assertFailedNum = 0
	)

	db, result, err, startTime, endTime, requestTime := client.SqlRequest(sqlInfo, sql.SqlString)
	defer db.Close()
	results := make(map[string]interface{})
	if sql.Debug == "all" {
		results["team_id"] = sql.TeamId
		results["sql_name"] = sql.Name
		results["target_id"] = sql.TargetId
		results["uuid"] = sql.Uuid.String()
		results["err"] = err
		results["request_time"] = requestTime / uint64(time.Millisecond)
		results["sql_result"] = result
		if err == nil {
			isSucceed = true
			results["status"] = []bool{isSucceed}
		} else {
			results["status"] = []bool{isSucceed}
		}
		by, _ := json.Marshal(sqlInfo)
		if by != nil {
			results["database"] = []string{string(by)}
		}
		results["sql"] = []string{sql.SqlString}
	}
	go model.Insert(mongoCollection, results, middlewares.LocalIp)
	return
}
