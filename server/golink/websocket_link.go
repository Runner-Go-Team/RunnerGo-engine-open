// Package golink 连接
package golink

import (
	"encoding/json"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/server/client"
	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/mongo"
)

func webSocketSend(ws model.WebsocketDetail, mongoCollection *mongo.Collection) (bool, int64, uint64, float64, float64) {
	var (
		// startTime = time.Now()
		isSucceed     = true
		errCode       = model.NoError
		receivedBytes = float64(0)
	)
	headers := map[string][]string{}
	for _, header := range ws.WsHeader {
		if header.IsChecked != model.Open {
			continue
		}
		headers[header.Var] = []string{header.Val}
	}
	//  api.Request.Body.ToString()
	recvResults, writeResults := make(map[string]interface{}), make(map[string]interface{})
	recvResults["type"] = "recv"
	recvResults["uuid"] = ws.Uuid.String()
	recvResults["name"] = ws.Name
	recvResults["team_id"] = ws.TeamId
	recvResults["target_id"] = ws.TargetId

	writeResults["type"] = "send"
	writeResults["uuid"] = ws.Uuid.String()
	writeResults["name"] = ws.Name
	writeResults["team_id"] = ws.TeamId
	writeResults["target_id"] = ws.TargetId
	header, _ := json.Marshal(headers)
	if header != nil {
		writeResults["header"] = string(header)
	} else {
		writeResults["header"] = ""
	}

	resp, requestTime, sendBytes, err := client.WebSocketRequest(recvResults, writeResults, mongoCollection, ws.Url, ws.SendMessage, headers, ws.WsConfig)

	if err != nil {
		isSucceed = false
		errCode = model.RequestError // 请求错误
	} else {
		// 接收到的字节长度
		receivedBytes, _ = decimal.NewFromFloat(float64(len(resp)) / 1024).Round(2).Float64()
	}
	return isSucceed, errCode, requestTime, float64(sendBytes), receivedBytes
}
