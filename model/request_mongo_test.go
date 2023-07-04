package model

import (
	"context"
	"fmt"
	"testing"
)

func TestQueryDebugStatus(t *testing.T) {

	mongoClient, err := NewMongoClient(
		"mongodb://runnergo:hello123456@192.168.1.205:27017/runnergo",
		"")
	if err != nil {
		fmt.Println("连接mongo错误：", err)
		return
	}
	defer mongoClient.Disconnect(context.TODO())

	fmt.Println("成功")
	//debugCollection := NewCollection("kunpeng", "debug_status", mongoClient)
	//filter := bson.D{{"report_id", 64}, {"plan_id", 29}, {"team_id", 197}}
	//m := make(map[string]interface{})
	//debugCollection.FindOne(context.TODO(), filter).Decode(m)
	//value, ok := m["debug"]
	//if ok {
	//	fmt.Println("123", value)
	//}

	//team := "197"
	//plan := "29"
	//report := "65"
	//fmt.Println(QueryDebugStatus(debugCollection, team, plan, report))

	//fmt.Println(QueryDebugStatus(debugCollection, 1298))

}
