package golink

//func SendDubbo(dubbo model.DubboDetail, mongoCollection *mongo.Collection) {
//	results := make(map[string]interface{})
//	results["uuid"] = dubbo.Uuid.String()
//	results["name"] = dubbo.Name
//	results["team_id"] = dubbo.TeamId
//	results["target_id"] = dubbo.TargetId
//	parameterTypes, parameterValues := []string{}, []hessian.Object{}
//
//	rpcServer, err := client.NewRpcServer(dubbo)
//
//	for _, parame := range dubbo.DubboParam {
//		if parame.IsChecked != constant.Open {
//			break
//		}
//		var val interface{}
//		switch parame.ParamType {
//		case constant.JavaInteger:
//			val, err = strconv.Atoi(parame.Val)
//			if err != nil {
//				val = parame
//				continue
//			}
//		case constant.JavaString:
//			val = parame.Val
//		case constant.JavaBoolean:
//			switch parame.Val {
//			case "true":
//				val = true
//			case "false":
//				val = false
//			default:
//				val = parame.Val
//			}
//		case constant.JavaByte:
//
//		case constant.JavaCharacter:
//		case constant.JavaDouble:
//			val, err = strconv.ParseFloat(parame.Val, 64)
//			if err != nil {
//				val = parame.Val
//				continue
//			}
//		case constant.JavaFloat:
//			val, err = strconv.ParseFloat(parame.Val, 64)
//			if err != nil {
//				val = parame.Val
//				continue
//			}
//			val = float32(val.(float64))
//		case constant.JavaLong:
//			val, err = strconv.ParseInt(parame.Val, 10, 64)
//			if err != nil {
//				val = parame.Val
//				continue
//			}
//		case constant.JavaMap:
//		case constant.JavaList:
//		default:
//			val = parame.Val
//		}
//		parameterTypes = append(parameterTypes, parame.ParamType)
//		parameterValues = append(parameterValues, val)
//
//	}
//	requestType, _ := json.Marshal(parameterTypes)
//	results["request_type"] = string(requestType)
//	requestBody, _ := json.Marshal(parameterValues)
//	results["request_body"] = string(requestBody)
//	if err != nil {
//		results["status"] = false
//		results["response_body"] = err.Error()
//	} else {
//		resp, err := rpcServer.(*generic.GenericService).Invoke(
//			context.TODO(),
//			dubbo.FunctionName,
//			parameterTypes,
//			parameterValues, // 实参
//		)
//		if err != nil {
//			results["status"] = false
//			results["response_body"] = err.Error()
//		}
//		if resp != nil {
//			response, _ := json.Marshal(resp)
//			results["status"] = true
//			results["response_body"] = string(response)
//
//		}
//
//	}
//	model.Insert(mongoCollection, results, middlewares.LocalIp)
//}
