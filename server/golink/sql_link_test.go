package golink

import (
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"testing"
)

func TestSqlSend(t *testing.T) {
	sqlInfo := model.SqlInfo{
		Type:     "mysql",
		Host:     "rm-2zem14s80lyu5c4z7.mysql.rds.aliyuncs.com",
		User:     "runnergo_open",
		Password: "czYNsm6LmfZ0XU3E",
		Port:     "3306",
		DB:       "runnergo_open",
	}
	sql := "select * from team where team_id = '154235dc-9aec-4eec-aa75-29a0f3770f21'"
	SqlSend("query", sql, sqlInfo)
}
