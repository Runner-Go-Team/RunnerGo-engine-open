package model

import (
	"context"
	"fmt"
	"github.com/lixiangyun/go-ntlm"
	"github.com/thedevsaddam/gojsonq"
	"testing"
)

func TestInsert(t *testing.T) {
	session, err := ntlm.CreateClientSession(ntlm.Version1, ntlm.ConnectionlessMode)
	if err != nil {
		return
	}
	session.SetUserInfo("auth.Ntlm.Username", "auth.Ntlm.Password", "auth.Ntlm.Domain")
	negotiate, err := session.GenerateNegotiateMessage()
	fmt.Println(negotiate)

}

func TestQueryPlanStatus(t *testing.T) {
	//rdb := redis.NewClient(
	//	&redis.Options{
	//		Addr:     "172.17.101.191:6398",
	//		Password: "apipost",
	//		DB:       0,
	//	})
	//_, err := rdb.Ping().Result()
	//value, err := rdb.Get("192.168.110.231:1934:90:1088:c03e1575-b5f0-450b-9c51-0c51f3c7ffd7:status").Result()
	//fmt.Println(err, "               ", value)

	ctx := context.Background()

	for i := 0; i < 10; i++ {
		go func(ctx1 context.Context, j int) {
			ctx1 = context.WithValue(ctx1, i, true)
			fmt.Println("ctx    ", ctx1.Value(i))
		}(ctx, i)
	}

	for i := 0; i < 10; i++ {
		fmt.Println(i, "     ", ctx.Value(i))
	}
}

func TestBody_SetBody(t *testing.T) {
	str := "{\n    \"scene_id\": 1415, \n    \"uuid\": \"0570c324-2bf3-4c33-b92d-8d8d7ab00f66\", \n    \"report_id\": 123, \n    \"team_id\": 158, \n    \"scene_name\": \"主流程\", \n    \"version\": 0, \n    \"debug\": \"\", \n    \"enable_plan_configuration\": false}"
	a := gojsonq.New().FromString(str)
	district := a.Find("report_id")
	fmt.Println(district)
}
