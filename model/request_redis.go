package model

import (
	"github.com/go-redis/redis"
	"time"
)

var (
	ReportRdb    *redis.Client
	RDB          *redis.Client
	timeDuration = 3 * time.Second
)

type RedisClient struct {
	Client *redis.Client
}

func InitRedisClient(reportAddr, reportPassword string, reportDb int64, addr, password string, db int64) (err error) {
	ReportRdb = redis.NewClient(
		&redis.Options{
			Addr:     reportAddr,
			Password: reportPassword,
			DB:       int(reportDb),
		})
	_, err = ReportRdb.Ping().Result()
	if err != nil {
		return err
	}

	RDB = redis.NewClient(
		&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       int(db),
		})
	_, err = RDB.Ping().Result()
	return err
}

func InsertStatus(key, value string, expiration time.Duration) (err error) {
	if expiration < 20*time.Second {
		expiration = 20 * time.Second
	}
	err = RDB.Set(key, value, expiration).Err()
	if err != nil {
		return
	}
	return
}

func QueryPlanStatus(key string) (err error, value string) {
	value, err = RDB.Get(key).Result()
	return
}

func QuerySceneStatus(key string) (err error, value string) {
	value, err = RDB.Get(key).Result()
	return
}

func QueryReportData(key string) (value string) {
	values := ReportRdb.LRange(key, 0, -1).Val()
	if len(values) <= 0 {
		return
	}
	value = values[0]
	return
}

func InsertHeartbeat(key string, field string, value interface{}) error {
	_, err := RDB.HSet(key, field, value).Result()
	return err
}

func DeleteKey(key string) (err error) {
	err = RDB.Del(key).Err()
	return
}

func InsertMachineResources(key string, value interface{}) error {
	_, err := RDB.LPush(key, value).Result()
	return err
}

func SubscribeMsg(topic string) (pubSub *redis.PubSub) {
	pubSub = RDB.Subscribe(topic)
	return
}
