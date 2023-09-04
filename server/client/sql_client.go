package client

import (
	"database/sql"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	_ "github.com/go-sql-driver/mysql" //mysql驱动
	_ "github.com/lib/pq"              //postgres驱动
)

func newMysqlClient(sqlInfo model.SqlDatabaseInfo) (db *sql.DB, err error) {
	var dsn string
	switch sqlInfo.Type {
	case "oracle":
		sqlInfo.Type = "oci8"
		dsn = fmt.Sprintf("%s:%s@%s:%d/%s?%s", sqlInfo.User, sqlInfo.Password, sqlInfo.Host, sqlInfo.Port, sqlInfo.DbName, sqlInfo.Charset)
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", sqlInfo.User, sqlInfo.Password, sqlInfo.Host, sqlInfo.Port, sqlInfo.DbName, sqlInfo.Charset)
	case "postgres":
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=verify-full", sqlInfo.User, sqlInfo.Password, sqlInfo.Host, sqlInfo.Port, sqlInfo.DbName)
	}

	db, err = sql.Open(sqlInfo.Type, dsn)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("%s数据库连接失败： %s", sqlInfo.Type, err.Error()))
		return
	}
	return
}

func TestConnection(sqlInfo model.SqlDatabaseInfo) (db *sql.DB, err error) {
	db, err = newMysqlClient(sqlInfo)
	if err != nil {
		return
	}
	err = db.Ping()
	if err != nil {
		return
	}
	return
}
