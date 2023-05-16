package golink

import (
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
	if result == nil {
		result = make(map[string]interface{})
	}
	if sql.Debug == "all" {
		result["uuid"] = sql.Uuid.String()
		result["err"] = err
		result["request_time"] = requestTime / uint64(time.Millisecond)
	}
	model.Insert(mongoCollection, result, middlewares.LocalIp)
	fmt.Println("time:     ", startTime, endTime)
}
