package model

import (
	"context"
	"fmt"
	"github.com/lixiangyun/go-ntlm"
	"net/http"
	"sync"
	"testing"
	"time"
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
	//		Addr:     ":6398",
	//		Password: "",
	//		DB:       0,
	//	})
	//_, err := rdb.Ping().Result()
	//value, err := rdb.Get(":90:1088:c03e1575-b5f0-450b-9c51-0c51f3c7ffd7:status").Result()
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

func Current(current int, duration time.Duration, ch chan bool, wg *sync.WaitGroup) {
	ticker := time.NewTicker(duration * time.Second)
	for i := 0; i <= current; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ticker.C:
					return
				default:
					req, _ := http.NewRequest("GET", "http://demo-api.runnergo.cn", nil)
					var client = new(http.Client)
					start := time.Now().UnixMilli()
					_, _ = client.Do(req)
					requestTime := time.Now().UnixMilli() - start
					fmt.Println(requestTime)
					//default:
					ch <- true

				}
			}
		}()

	}
}

func TestBody_SetBody(t *testing.T) {
	//str := "{\n    \"scene_id\": 1415, \n    \"uuid\": \"0570c324-2bf3-4c33-b92d-8d8d7ab00f66\", \n    \"report_id\": 123, \n    \"team_id\": 158, \n    \"scene_name\": \"主流程\", \n    \"version\": 0, \n    \"debug\": \"\", \n    \"enable_plan_configuration\": false}"
	//a := gojsonq.New().FromString(str)
	//district := a.Find("report_id")
	//fmt.Println(district)

	var wg = new(sync.WaitGroup)
	wg.Add(1)
	var num = 0
	str := time.Now().UnixMilli()
	var ch = make(chan bool, 10000)
	Current(200, 30, ch, wg)
	wg.Wait()
	en := time.Now().UnixMilli() - str
	fmt.Println("sq : ", en, "    ", num)
	for {
		select {
		case i := <-ch:
			if i {
				num++
				fmt.Println("sq : ", en, "    ", num)
			}
		}
	}

}
