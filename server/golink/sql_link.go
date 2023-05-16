package golink

import (
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/server/client"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

func SqlSend(sql model.SQL, sqlInfo model.MysqlDatabaseInfo, mongoCollection *mongo.Collection) {
	//var (
	//	isSucceed       = true
	//	errCode         = model.NoError
	//	receivedBytes   = float64(0)
	//	errMsg          = ""
	//	assertNum       = 0
	//	assertFailedNum = 0
	//)

	db, result, err, startTime, endTime, requestTime := client.SqlRequest(sqlInfo, sql.SqlString)
	defer db.Close()
	results := make(map[string]interface{})
	if sql.Debug == "all" {
		results["uuid"] = sql.Uuid.String()
		results["err"] = err
		results["request_time"] = requestTime / uint64(time.Millisecond)
		results["sql_result"] = result
		if err != nil {
			results["status"] = []string{"success"}
		} else {
			results["status"] = []string{"failed"}
		}
		by, _ := json.Marshal(sqlInfo)
		if by != nil {
			results["database"] = []string{string(by)}
		}
		results["sql"] = []string{sql.SqlString}
	}
	model.Insert(mongoCollection, results, middlewares.LocalIp)
	fmt.Println("time:     ", startTime, endTime)
}
