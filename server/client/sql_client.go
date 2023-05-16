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

func SqlRequest(sqlInfo model.MysqlDatabaseInfo, sqls string) (db *sql.DB, resultMap map[string][]string, err error, startTime, endTime time.Time, requestTime uint64) {
	db, err = newMysqlClient(sqlInfo)
	if db == nil || err != nil {
		return
	}
	startTime = time.Now()
	if strings.HasPrefix(sqls, "select") {
		rows, errQuery := db.Query(sqls)
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
	} else {
		result, errExec := db.Exec(sqls)
		if errExec != nil || result == nil {
			err = errExec
			return
		}
		row, errExec := result.RowsAffected()
		if errExec != nil {
			fmt.Println("row err :   ", errExec)
		}
		last, errExec := result.LastInsertId()
		if errExec != nil {
			fmt.Println("last err :   ", errExec)
		}
		fmt.Println("row:   ", row)
		fmt.Println("last:   ", last)
	}
	endTime = time.Now()
	requestTime = uint64(time.Since(startTime))
	return
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
