package js

type JS struct {
	Str string `json:"str"`
}

//
//func RunJs(script string) (value *v8go.Value) {
//	//defer tools.DeferPanic("js脚本格式错误")
//	//ctx := v8go.NewContext()
//	//value, err := ctx.RunScript(script, "main.js")
//	//if err != nil {
//	//	log.Logger.Error("js脚本运行错误：", err)
//	//	return nil
//	//}
//	return
//}
