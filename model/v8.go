package model

import "rogchap.com/v8go"

type JS struct {
	Str string `json:"str"`
}

func RunJs() {
	ctx := v8go.NewContext()
	ctx.RunScript("console.log(123)", "")
}
