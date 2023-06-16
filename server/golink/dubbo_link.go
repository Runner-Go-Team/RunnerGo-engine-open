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
	"strconv"
)

func SendDubbo(dubbo model.DubboDetail, mongoCollection *mongo.Collection) {
	results := make(map[string]interface{})
	results["uuid"] = dubbo.Uuid.String()
	results["name"] = dubbo.Name
	results["team_id"] = dubbo.TeamId
	results["target_id"] = dubbo.TargetId
	parameterTypes, parameterValues := []string{}, []hessian.Object{}

	rpcServer, err := client.NewRpcServer(dubbo)
	for _, parame := range dubbo.DubboParam {
		if parame.IsChecked != model.Open {
			break
		}
		var val interface{}
		switch parame.ParamType {
		case model.JavaInteger:
			val, err = strconv.Atoi(parame.Val)
			if err != nil {
				val = parame
				continue
			}
		case model.JavaString:
			val = parame.Val
		case model.JavaBoolean:
			switch parame.Val {
			case "true":
				val = true
			case "false":
				val = false
			default:
				val = parame.Val
			}
		case model.JavaByte:

		case model.JavaCharacter:
		case model.JavaDouble:
			val, err = strconv.ParseFloat(parame.Val, 64)
			if err != nil {
				val = parame.Val
				continue
			}
		case model.JavaFloat:
			val, err = strconv.ParseFloat(parame.Val, 64)
			if err != nil {
				val = parame.Val
				continue
			}
			val = float32(val.(float64))
		case model.JavaLong:
			val, err = strconv.ParseInt(parame.Val, 10, 64)
			if err != nil {
				val = parame.Val
				continue
			}
		case model.JavaMap:
		case model.JavaList:
		default:
			val = parame.Val
		}
		parameterTypes = append(parameterTypes, parame.ParamType)
		parameterValues = append(parameterValues, val)

	}
	requestType, _ := json.Marshal(parameterTypes)
	results["request_type"] = string(requestType)
	requestBody, _ := json.Marshal(parameterValues)
	results["request_body"] = string(requestBody)

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
