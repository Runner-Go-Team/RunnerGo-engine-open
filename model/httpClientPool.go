package model

import "github.com/valyala/fasthttp"

var WorkPool []*fasthttp.Client

func CreatWorkPool(num int) {
	for i := 0; i < num; i++ {

	}
}
