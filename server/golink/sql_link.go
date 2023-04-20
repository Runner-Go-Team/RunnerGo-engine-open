package golink

import (
	"database/sql"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/server/client"
)

func SqlSend(action, sqls string, sqlInfo model.SqlInfo) {
	//var (
	//	isSucceed       = true
	//	errCode         = model.NoError
	//	receivedBytes   = float64(0)
	//	errMsg          = ""
	//	assertNum       = 0
	//	assertFailedNum = 0
	//)

	rows, result, err, startTime, endTime, requestTime := client.SqlRequest(sqlInfo, action, sqls)
	switch action {

	case "query":
		if rows == nil {
			return
		}
		defer rows.Close()
		cols, _ := rows.Columns()
		values := make([]sql.RawBytes, len(cols))
		scans := make([]interface{}, len(cols))
		if cols != nil {
			for i := range cols {
				scans[i] = &values[i]
			}
		}
		fmt.Println("cols:  ", cols)
		results := make(map[string][]string)
		for rows.Next() {
			if err := rows.Scan(scans...); err != nil {
				continue
			}
			for j, v := range values {
				results[cols[j]] = append(results[cols[j]], string(v))
			}
		}
		fmt.Println("results:    ", results)
	default:
		if result == nil {
			return
		}
		row, err := result.RowsAffected()
		if err != nil {
			fmt.Println("row err :   ", err)
		}
		last, err := result.LastInsertId()
		if err != nil {
			fmt.Println("last err :   ", err)
		}
		fmt.Println("row:   ", row)
		fmt.Println("last:   ", last)

	}

	fmt.Println("err:   ", err)
	fmt.Println("startTime:   ", startTime)
	fmt.Println("endTime:   ", endTime)
	fmt.Println("requestTime:   ", requestTime)
}
