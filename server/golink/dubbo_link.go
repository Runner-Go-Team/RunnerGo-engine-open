package golink

import (
	"context"
	"dubbo.apache.org/dubbo-go/v3/config/generic"
	"encoding/json"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/server/client"
	hessian "github.com/apache/dubbo-go-hessian2"
	"go.mongodb.org/mongo-driver/mongo"
)

func SendDubbo(dubbo model.DubboDetail, mongoCollection *mongo.Collection) {
	results := make(map[string]interface{})
	results["uuid"] = dubbo.Uuid.String()
	results["name"] = dubbo.Name
	results["team_id"] = dubbo.TeamId
	results["target_id"] = dubbo.TargetId
	parameterTypes, parameterValues := []string{}, []hessian.Object{}
	requestType, _ := json.Marshal(parameterTypes)
	results["request_type"] = string(requestType)
	requestBody, _ := json.Marshal(parameterValues)
	results["request_body"] = string(requestBody)

	rpcServer, err := client.NewRpcServer(dubbo)

	for _, parame := range dubbo.DubboParam {
		if parame.IsChecked != model.Open {
			break
		}
		parameterTypes = append(parameterTypes, parame.ParamType)
		parameterValues = append(parameterValues, parame.Val)
	}
	if err != nil {
		results["response_body"] = err.Error()
	} else {

		resp, err := rpcServer.(*generic.GenericService).Invoke(
			context.TODO(),
			dubbo.FunctionName,
			parameterTypes,
			parameterValues, // 实参
		)
		if err != nil {
			results["response_body"] = err.Error()
		}
		if resp != nil {
			response, _ := json.Marshal(resp)
			results["response_body"] = string(response)
		}

	}

	model.Insert(mongoCollection, results, middlewares.LocalIp)
}
