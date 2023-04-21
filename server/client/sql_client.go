package client

import (
	"database/sql"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	_ "github.com/go-sql-driver/mysql" //mysql驱动
	_ "github.com/lib/pq"              //postgres驱动
	"time"
)

func SqlRequest(sqlInfo model.SqlInfo, action, sqls string) (resultMap map[string][]string, err error, startTime, endTime time.Time, requestTime uint64) {
	db := newMysqlClient(sqlInfo)
	if db == nil {
		return
	}
	startTime = time.Now()
	defer db.Close()
	switch action {
	case "query":
		rows, err := db.Query(sqls)
		if err != nil || rows == nil {
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
		results := make(map[string][]string)
		for rows.Next() {
			if err := rows.Scan(scans...); err != nil {
				continue
			}
			for j, v := range values {
				results[cols[j]] = append(results[cols[j]], string(v))
			}
		}
		for k, value := range results {
			fmt.Println(k, "    ", value)
		}
	default:
		result, err := db.Exec(sqls)
		if err != nil || result == nil {
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
	endTime = time.Now()
	requestTime = uint64(time.Since(startTime))
	return
}

func newMysqlClient(sqlInfo model.SqlInfo) (db *sql.DB) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", sqlInfo.User, sqlInfo.Password, sqlInfo.Host, sqlInfo.Port, sqlInfo.DB, sqlInfo.Charset)
	db, err := sql.Open(sqlInfo.Type, dsn)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("%s数据库连接失败： %s", sqlInfo.Type, err.Error()))
		return
	}
	return
}
