package tools

import (
	"github.com/valyala/fasthttp"
	"strings"
)

type Condition struct {
	Code string
}

func IfController(response *fasthttp.Response, conditions []Condition) {
	if response == nil || conditions == nil {
		return
	}

	for _, condition := range conditions {
		if condition.Code == response.String() {

		}
	}
}

// 判断分隔符 = in is :

func BreakUp(str string, symbol string) (s, result string) {
	if str == "" || symbol == "" {
		return
	}
	results := strings.Split(str, symbol)
	s = strings.TrimSpace(results[0])
	result = strings.TrimSpace(results[1])
	return
}
