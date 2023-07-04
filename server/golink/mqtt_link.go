package golink

//
//func SendMqtt(mqtt model.MQTT, mongoCollection *mongo.Collection) {
//	results := make(map[string]interface{})
//	results["uuid"] = mqtt.Uuid.String()
//	results["name"] = mqtt.Name
//	results["team_id"] = mqtt.TeamId
//	results["target_id"] = mqtt.TargetId
//	mqttClient, err := client.NewMqttClient(mqtt.MQTTConfig)
//	if err != nil {
//		results["err"] = err.Error()
//		results["request_body"] = ""
//		model.Insert(mongoCollection, results, middlewares.LocalIp)
//	}
//	mqttClient.Client.Disconnect()
//}
