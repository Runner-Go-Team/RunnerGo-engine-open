package js

import (
	"testing"
)

func add(a, b int) (c int, s string) {
	s = "1231232123"
	return a + b, s
}
func TestRunJs(t *testing.T) {
	//m := new(sync.Map)
	//m.Store("name", 10)
	//m.Store("age", 30)
	//js := "let a = 10\nlet b = 20\n\nfunction add(a, b) {\n console.log(121231231231) \n return [a+b , 50]  \n}\n\n \n function ad(b) {\n return [111, b]\n}\n\nlet c = add({{name}}, {{age}})\n\nad(c[1]-c[0])"
	//results := tools.FindAllDestStr(js, "{{(.*?)}}")
	//if results == nil || len(results) <= 0 {
	//	return
	//}
	//for _, result := range results {
	//	if result == nil || len(result) <= 1 {
	//		continue
	//	}
	//	if value, ok := m.Load(result[1]); ok {
	//		value = fmt.Sprintf("%d", value)
	//		js = strings.Replace(js, result[0], value.(string), -1)
	//	}
	//}
	//RunJs(js)
}
