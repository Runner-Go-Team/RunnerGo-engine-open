package golink

import (
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/server/client"
)

func SqlSend(action, sqls string, sqlInfo model.MysqlDatabaseInfo) {
	//var (
	//	isSucceed       = true
	//	errCode         = model.NoError
	//	receivedBytes   = float64(0)
	//	errMsg          = ""
	//	assertNum       = 0
	//	assertFailedNum = 0
	//)
	db, result, err, startTime, endTime, requestTime := client.SqlRequest(sqlInfo, action, sqls)
	defer db.Close()
	fmt.Println("result:   ", result)
	fmt.Println("err:   ", err)
	fmt.Println("startTime:   ", startTime)
	fmt.Println("endTime:   ", endTime)
	fmt.Println("requestTime:   ", requestTime)
}
