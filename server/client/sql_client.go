package client

import (
	"database/sql"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	_ "github.com/go-sql-driver/mysql" //mysql驱动
	_ "github.com/lib/pq"              //postgres驱动
	"strings"
	"time"
)

func SqlRequest(sqlInfo model.MysqlDatabaseInfo, sqls string) (db *sql.DB, result map[string]interface{}, err error, startTime, endTime time.Time, requestTime uint64) {
	db, err = newMysqlClient(sqlInfo)
	if db == nil || err != nil {
		return
	}

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
		values := make([]sql.RawBytes, len(cols))
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
			log.Logger.Error("row err :   ", row)
		}
		result["rows_affected"] = row
		last, errExec := results.LastInsertId()
		if errExec != nil {
			log.Logger.Error("last err :   ", errExec)
		}
		result["last_insert_id"] = last
		return
	}

}

func newMysqlClient(sqlInfo model.MysqlDatabaseInfo) (db *sql.DB, err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", sqlInfo.User, sqlInfo.Password, sqlInfo.Host, sqlInfo.Port, sqlInfo.DbName, sqlInfo.Charset)
	db, err = sql.Open(sqlInfo.Type, dsn)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("%s数据库连接失败： %s", sqlInfo.Type, err.Error()))
		return
	}
	return
}

func TestConnection(sqlInfo model.MysqlDatabaseInfo) (db *sql.DB, err error) {
	return newMysqlClient(sqlInfo)
}
