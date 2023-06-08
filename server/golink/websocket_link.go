// Package golink 连接
package golink

import (
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

	resp, requestTime, sendBytes, err := client.WebSocketRequest(ws.Url, ws.SendMessage, headers, ws.WsConfig)

	if err != nil {
		isSucceed = false
		errCode = model.RequestError // 请求错误
	} else {
		// 接收到的字节长度
		receivedBytes, _ = decimal.NewFromFloat(float64(len(resp)) / 1024).Round(2).Float64()
	}
	return isSucceed, errCode, requestTime, float64(sendBytes), receivedBytes
}
