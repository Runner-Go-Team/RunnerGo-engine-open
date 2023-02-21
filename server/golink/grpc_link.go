// Package golink 连接
package golink

//
//// Grpc grpc 接口请求
//func Grpc(chanID uint64, ch chan<- *model.ResultDataMsg, totalNumber uint64, wg *sync.WaitGroup,
//	api *model.Api, ws *client.GrpcSocket) {
//	defer func() {
//		wg.Done()
//	}()
//	defer func() {
//		_ = ws.Close()
//	}()
//	for i := uint64(0); i < totalNumber; i++ {
//		grpcRequest(chanID, ch, i, api, ws)
//	}
//	return
//}

//// grpcRequest 请求
//func grpcRequest(chanID uint64, ch chan<- *model.ResultDataMsg, i uint64, api *model.Api,
//	ws *client.GrpcSocket) {
//	var (
//		startTime = time.Now().UnixMilli()
//		isSucceed = false
//		errCode   = model.NoError
//	)
//	// 需要发送的数据
//	conn := ws.GetConn()
//	if conn == nil {
//		errCode = model.RequestError
//	} else {
//		// TODO::请求接口示例
//		c := pb.NewApiServerClient(conn)
//		var (
//			ctx = context.Background()
//			req = &pb.Request{
//				//UserName: api.Request.Body,
//			}
//		)
//		rsp, err := c.HelloWorld(ctx, req)
//		// fmt.Printf("rsp:%+v", rsp)
//		if err != nil {
//			errCode = model.RequestError
//		} else {
//			// 200 为成功
//			if rsp.Code != 200 {
//				errCode = model.RequestError
//			} else {
//				isSucceed = true
//			}
//		}
//	}
//	requestTime := tools.TimeDifference(startTime)
//	requestResults := &model.ResultDataMsg{
//		RequestTime: requestTime,
//		IsSucceed:   isSucceed,
//		ErrorType:   errCode,
//	}
//	ch <- requestResults
//}
