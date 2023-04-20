package client

import (
	"database/sql"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"time"
)

func SqlRequest(sqlInfo model.SqlInfo, action, sql string) (rows *sql.Rows, result sql.Result, err error, startTime, endTime time.Time, requestTime uint64) {
	db := newMysqlClient(sqlInfo.Type, sqlInfo.User, sqlInfo.Password, sqlInfo.Host, sqlInfo.Port, sqlInfo.DB, sqlInfo.Charset)
	if db == nil {
		return
	}
	startTime = time.Now()
	defer db.Close()
	switch action {
	case "query":
		rows, err = db.Query(sql)
	default:
		result, err = db.Exec(sql)
	}
	endTime = time.Now()
	requestTime = uint64(time.Since(startTime))
	return
}

func newMysqlClient(sqlType, user, password, host, port, dbName, charset string) (db *sql.DB) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", user, password, host, port, dbName, charset)
	db, err := sql.Open(sqlType, dsn)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("%s数据库连接失败： %s", sqlType, err.Error()))
		return
	}
	return
}
