package model

import (
	"context"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/constant"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewMongoClient(dsn, ip string) (mongoClient *mongo.Client, err error) {
	clientOptions := options.Client().ApplyURI(dsn)
	mongoClient, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return
	}

	err = mongoClient.Ping(context.TODO(), nil)
	if err != nil {
		return
	}
	//log.Logger.Info(fmt.Sprintf("机器ip:%s, mongo数据库建连成功!", ip))
	return
}

func NewCollection(db, table string, mongoClient *mongo.Client) (collection *mongo.Collection) {
	collection = mongoClient.Database(db).Collection(table)
	return
}

func Insert(collection *mongo.Collection, msg interface{}, ip string) {
	_, err := collection.InsertOne(context.TODO(), msg)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("机器ip:%s, 向mongo写入数据错误:  %s", ip, err.Error()))
	}
}

// QueryDebugStatus 查询debug状态
func QueryDebugStatus(collection *mongo.Collection, teamId, planId, reportId string) string {
	filter := bson.D{{"report_id", reportId}, {"team_id", teamId}, {"plan_id", planId}}
	m := make(map[string]interface{})
	_ = collection.FindOne(context.TODO(), filter).Decode(m)

	value, ok := m["debug"]
	if ok {
		return value.(string)
	}
	return constant.StopDebug
}
