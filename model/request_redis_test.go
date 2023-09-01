package model

import (
	"fmt"
	"testing"
)

func TestSubscribeMsg(t *testing.T) {
	//address := ":6398"
	//password := ""
	//
	//rdb := redis.NewClient(
	//	&redis.Options{
	//		Addr:     address,
	//		Password: password,
	//		DB:       0,
	//	})
	//wg := new(sync.WaitGroup)
	//ch := rdb.Subscribe("SubscriptionStressPlanStatusChange:242:1:21").Channel()
	//for i := 0; i < 1; i++ {
	//	wg.Add(1)
	//	go func() {
	//		for {
	//			select {
	//			case msg := <-ch:
	//				fmt.Println("channel:    ", msg.Channel)
	//				fmt.Println("payload:    ", msg.Payload)
	//				fmt.Println("string:     ", msg.String())
	//			default:
	//				fmt.Println(1231231312)
	//				time.Sleep(2 * time.Second)
	//			}
	//
	//		}
	//	}()
	//}
	//time.Sleep(5 * time.Second)
	//wg.Add(1)
	//go func() {
	//	a, err := rdb.Publish("topic", "123").Result()
	//	fmt.Println("err:  ", err, "      ", a)
	//}()
	//wg.Wait()
	//a := 1
	//for a < 9999999999999 {
	//	a++
	//}

	var a interface{}
	fmt.Println("a :    ", a)

}
