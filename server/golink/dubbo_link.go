package golink

import (
	"context"
	"dubbo.apache.org/dubbo-go/v3/config/generic"
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
	rpcServer, err := client.NewRpcServer(dubbo)
	if err != nil {
		results["err"] = err
	} else {
		parameterTypes, parameterValues := []string{}, []hessian.Object{}

		for _, parame := range dubbo.DubboParam {
			if parame.IsChecked != model.Open {
				break
			}
			parameterTypes = append(parameterTypes, parame.ParamType)
			parameterValues = append(parameterValues, parame.Val)
		}

		results["request_type"] = parameterTypes
		results["request_body"] = parameterTypes
		resp, err := rpcServer.(*generic.GenericService).Invoke(
			context.TODO(),
			dubbo.FunctionName,
			parameterTypes,
			parameterValues, // 实参
		)
		results["err"] = err
		results["response_body"] = resp
	}

	model.Insert(mongoCollection, results, middlewares.LocalIp)
}
