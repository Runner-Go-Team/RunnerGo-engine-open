package golink

import (
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/server/client"
	"go.mongodb.org/mongo-driver/mongo"
)

func SendMqtt(mqtt model.MQTT, mongoCollection *mongo.Collection) {
	results := make(map[string]interface{})
	results["uuid"] = mqtt.Uuid.String()
	results["name"] = mqtt.Name
	results["team_id"] = mqtt.TeamId
	results["target_id"] = mqtt.TargetId
	_, err := client.NewMqttClient(mqtt.MQTTConfig)
	if err != nil {
		results["err"] = err.Error()
		results["request_body"] = ""
		model.Insert(mongoCollection, results, middlewares.LocalIp)
	}
}
