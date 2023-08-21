package model

import (
	"github.com/go-redis/redis"
	"strings"
	"time"
)

var (
	RDB          *redis.ClusterClient
	timeDuration = 3 * time.Second
)

type RedisClient struct {
	Client *redis.ClusterClient
}

func InitRedisClient(clusterAddr, password string) (err error) {
	RDB = redis.NewClusterClient(
		&redis.ClusterOptions{
			Addrs:    strings.Split(clusterAddr, ";"),
			Password: password,
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
	values := RDB.LRange(key, 0, -1).Val()
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
