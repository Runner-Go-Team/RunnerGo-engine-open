package model

import (
	"RunnerGo-engine/tools"
	"fmt"
	"github.com/valyala/fasthttp"
	"strconv"
	"strings"
)

// Assertion 断言
type Assertion struct {
	Type             int8              `json:"type"` //  0:Text; 1:Regular; 2:Json; 3:XPath
	AssertionText    *AssertionText    `json:"assertionText"`
	AssertionRegular *AssertionRegular `json:"assertionRegular"`
	AssertionJson    *AssertionJson    `json:"assertionJson"`
	AssertionXPath   *AssertionXPath   `json:"assertionXPath"`
}

// AssertionText 文本断言 0
type AssertionText struct {
	IsChecked    int    `json:"is_checked"`    // 1 选中  2 未选
	ResponseType int8   `json:"response_type"` //  1:ResponseHeaders; 2:ResponseData; 3: ResponseCode;
	Compare      string `json:"compare"`       // Includes、UNIncludes、Equal、UNEqual、GreaterThan、GreaterThanOrEqual、LessThan、LessThanOrEqual、Includes、UNIncludes、NULL、NotNULL、OriginatingFrom、EndIn
	Var          string `json:"var"`           // 变量
	Val          string `json:"val"`           // 值
}

// AssertionRegular 正则断言 1
type AssertionRegular struct {
	AssertionTarget int8   `json:"type"`       // 2:ResponseData
	Expression      string `json:"expression"` // 正则表达式

}

// AssertionJson json断言 2
type AssertionJson struct {
	Expression string `json:"expression"` // json表达式
	Condition  string `json:"condition"`  // Contain、NotContain、Equal、NotEqual
}

// AssertionXPath xpath断言 3
type AssertionXPath struct {
}

// VerifyAssertionText 验证断言 文本断言
func (assertionText *AssertionText) VerifyAssertionText(response *fasthttp.Response) (code int64, ok bool, msg string) {
	assertionText.Var = strings.TrimSpace(assertionText.Var)
	assertionText.Val = strings.TrimSpace(assertionText.Val)
	switch assertionText.ResponseType {
	case ResponseCode:
		value, err := strconv.Atoi(assertionText.Val)
		if err != nil {
			return AssertError, false, assertionText.Val + "不是int类型,转换失败"
		}
		switch assertionText.Compare {
		case Equal:
			if value == response.StatusCode() {
				return NoError, true, fmt.Sprintf("响应码 等于 %d, 断言：成功！", value)
			} else {
				return AssertError, false, fmt.Sprintf("响应码 等于 %d, 断言：失败！", value)
			}
		case UNEqual:
			if value != response.StatusCode() {
				return NoError, true, fmt.Sprintf("响应码 不等于 %d断言：成功！", value)
			} else {
				return AssertError, false, fmt.Sprintf("响应码 不等于 %d断言：失败！", value)
			}
		default:
			return AssertError, false, "响应码断言条件不正确！"
		}
	case ResponseHeaders:
		header := response.Header.String()
		switch assertionText.Compare {
		case Includes:
			if strings.Contains(header, assertionText.Val) {
				return NoError, true, "响应头中包含：" + assertionText.Val + " 断言: 成功！"
			} else {
				return AssertError, false, "响应头中包含：" + assertionText.Val + " 断言: 失败！"
			}
		case NULL:
			if response.Header.String() == "" {
				return NoError, true, "响应头为空，断言: 成功！"
			} else {
				return AssertError, false, "响应头为空， 断言: 失败！"
			}
		case UNIncludes:
			if strings.Contains(header, assertionText.Val) {
				return AssertError, false, "响应头中不包含：" + assertionText.Val + " 断言: 失败！"
			} else {
				return NoError, true, "响应头中不包含：" + assertionText.Val + " 断言: 成功！"

			}
		case NotNULL:
			if header != "" {
				return NoError, true, "响应头不为空， 断言: 成功！"
			} else {
				return AssertError, false, "响应头不为空， 断言: 失败！"
			}
		default:
			return AssertError, false, "Header断言条件不正确！"
		}
	case ResponseData:
		resp := string(response.Body())
		switch assertionText.Compare {
		case Equal:
			value := tools.JsonPath(resp, assertionText.Var)
			if value == nil {
				return AssertError, false, "响应中" + assertionText.Var + "不存在， 断言: 失败！"
			}
			switch fmt.Sprintf("%T", value) {
			case "string":
				if value == assertionText.Val {
					return NoError, true, fmt.Sprintf("%s 等于 %s, 断言： 成功！", assertionText.Var, value)
				} else {
					return AssertError, false, fmt.Sprintf("%s 等于 %s, 断言： 失败！", assertionText.Var, value)
				}
			case "bool":
				if value == true && assertionText.Val == "true" {
					return NoError, true, fmt.Sprintf("%s 等于 %t, 断言： 成功！", assertionText.Var, value)
				} else if value == false && assertionText.Val == "false" {
					return NoError, true, fmt.Sprintf("%s 等于 %t, 断言： 成功！", assertionText.Var, value)
				} else {
					return AssertError, false, fmt.Sprintf("%s 等于 %t, 断言： 失败！", assertionText.Var, value)
				}
			case "float64":
				if num, err := strconv.ParseFloat(assertionText.Val, 64); err == nil {
					if value == num {
						return NoError, true, fmt.Sprintf("%s 等于 %s, 断言： 成功！", assertionText.Var, assertionText.Val)
					} else {
						return AssertError, false, fmt.Sprintf("%s 等于 %s, 断言： 失败！", assertionText.Var, assertionText.Val)
					}
				} else {
					return AssertError, false, "不是数字类型，无法比较大小！"
				}
			}

		case UNEqual:
			value := tools.JsonPath(resp, assertionText.Var)
			if value == nil {
				return AssertError, false, "响应中" + assertionText.Var + "不存在， 断言: 失败！"
			}
			switch fmt.Sprintf("%T", value) {
			case "string":
				if value != assertionText.Val {
					return NoError, true, fmt.Sprintf("%s 不等于 %s, 断言： 成功！", assertionText.Var, assertionText.Val)
				} else {
					return AssertError, false, fmt.Sprintf("%s 不等于 %s, 断言： 失败！", assertionText.Var, assertionText.Val)
				}
			case "bool":
				if value != true && assertionText.Val != "true" {
					return NoError, true, fmt.Sprintf("%s 不等于 %s, 断言： 成功！", assertionText.Var, assertionText.Val)
				} else if value != false && assertionText.Val != "false" {
					return NoError, true, fmt.Sprintf("%s 不等于 %s, 断言： 成功！", assertionText.Var, assertionText.Val)
				} else {
					return AssertError, false, fmt.Sprintf("%s 不等于 %s, 断言： 失败！", assertionText.Var, assertionText.Val)
				}
			case "float64":
				if num, err := strconv.ParseFloat(assertionText.Val, 64); err == nil {
					if value != num {
						return NoError, true, fmt.Sprintf("%s 不等于 %s, 断言： 成功！", assertionText.Var, assertionText.Val)
					} else {
						return AssertError, false, fmt.Sprintf("%s 不等于 %s, 断言： 失败！", assertionText.Var, assertionText.Val)
					}
				} else {
					return AssertError, false, "不是数字类型，无法比较大小！"
				}
			}
		case Includes:
			if strings.Contains(resp, assertionText.Val) {
				return NoError, true, "响应体中包含：" + assertionText.Val + " 断言: 成功！"
			} else {
				return AssertError, false, "响应体中包含：" + assertionText.Val + " 断言: 失败！"
			}
		case UNIncludes:
			if strings.Contains(resp, assertionText.Val) {
				return AssertError, false, "响应体中不包含：" + assertionText.Val + " 断言: 失败！"
			} else {
				return NoError, true, "响应体中不包含：" + assertionText.Val + " 断言: 成功！"
			}
		case NULL:
			if resp == "" {
				return NoError, true, "响应体为空， 断言: 成功！"
			} else {
				return AssertError, false, "响应体为空， 断言: 失败！"
			}
		case NotNULL:
			if resp == "" {
				return AssertError, false, "响应体不为空， 断言: 失败！"
			} else {
				return NoError, true, "响应体不为空， 断言: 成功！"
			}

		case GreaterThan:
			value := tools.JsonPath(resp, assertionText.Var)
			if value == nil {
				return AssertError, false, "响应体中" + assertionText.Var + "不存在， 断言: 失败！"
			}
			if num, err := strconv.ParseFloat(assertionText.Val, 64); err == nil {
				switch fmt.Sprintf("%T", value) {
				case "string":
					if i, err := strconv.ParseFloat(value.(string), 64); err == nil {
						if i > num {
							return NoError, true, fmt.Sprintf("%s 大于 %s, 断言： 成功！", assertionText.Var, assertionText.Val)
						} else {
							return AssertError, false, fmt.Sprintf("%s 大于 %s, 断言： 失败！", assertionText.Var, assertionText.Val)
						}
					} else {
						return AssertError, false, "不是数字类型，无法比较大小！"
					}
				case "float64":
					if value.(float64) > num {
						return NoError, true, fmt.Sprintf("%s 大于 %s, 断言： 成功！", assertionText.Var, assertionText.Val)
					} else {
						return AssertError, false, fmt.Sprintf("%s 大于 %s, 断言： 失败！", assertionText.Var, assertionText.Val)
					}
				}

			} else {
				return AssertError, false, "不是数字类型，无法比较大小！"
			}
		case GreaterThanOrEqual:
			value := tools.JsonPath(resp, assertionText.Var)
			if value == nil {
				return AssertError, false, "响应体中" + assertionText.Var + "不存在， 断言: 失败！"
			}
			if num, err := strconv.ParseFloat(assertionText.Val, 64); err == nil {
				switch fmt.Sprintf("%T", value) {
				case "string":
					if i, err := strconv.ParseFloat(value.(string), 64); err == nil {
						if i >= num {
							return NoError, true, fmt.Sprintf("%s 大于等于 %s, 断言： 成功！", assertionText.Var, assertionText.Val)
						} else {
							return AssertError, false, fmt.Sprintf("%s 大于等于 %s, 断言： 失败！", assertionText.Var, assertionText.Val)
						}
					} else {
						return AssertError, false, "不是数字类型，无法比较大小！"
					}
				case "float64":
					if value.(float64) >= num {
						return NoError, true, fmt.Sprintf("%s 大于等于 %s, 断言： 成功！", assertionText.Var, assertionText.Val)
					} else {
						return AssertError, false, fmt.Sprintf("%s 大于等于 %s, 断言： 失败！", assertionText.Var, assertionText.Val)
					}
				}

			} else {
				return AssertError, false, "不是数字类型，无法比较大小！"
			}

		case LessThan:
			value := tools.JsonPath(resp, assertionText.Var)
			if value == nil {
				return AssertError, false, "响应体中" + assertionText.Var + "不存在， 断言: 失败！"
			}
			if num, err := strconv.ParseFloat(assertionText.Val, 64); err == nil {
				switch fmt.Sprintf("%T", value) {
				case "string":
					if i, err := strconv.ParseFloat(value.(string), 64); err == nil {
						if i < num {
							return NoError, true, fmt.Sprintf("%s 小于 %s, 断言： 成功！", assertionText.Var, assertionText.Val)
						} else {
							return AssertError, false, fmt.Sprintf("%s 小于 %s, 断言： 失败！", assertionText.Var, assertionText.Val)
						}
					} else {
						return AssertError, false, "不是数字类型，无法比较大小！"
					}
				case "float64":
					if value.(float64) < num {
						return NoError, true, fmt.Sprintf("%s 小于 %s, 断言： 成功！", assertionText.Var, assertionText.Val)
					} else {
						return AssertError, false, fmt.Sprintf("%s 小于 %s, 断言： 失败！", assertionText.Var, assertionText.Val)
					}
				}

			} else {
				return AssertError, false, "不是数字类型，无法比较大小！"
			}
		case LessThanOrEqual:
			value := tools.JsonPath(resp, assertionText.Var)
			if value == nil {
				return AssertError, false, "响应体中" + assertionText.Var + "不存在， 断言: 失败！"
			}
			if num, err := strconv.ParseFloat(assertionText.Val, 64); err == nil {
				switch fmt.Sprintf("%T", value) {
				case "string":
					if i, err := strconv.ParseFloat(value.(string), 64); err == nil {
						if i <= num {
							return NoError, true, fmt.Sprintf("%s 小于等于 %s, 断言： 成功！", assertionText.Var, assertionText.Val)
						} else {
							return AssertError, false, fmt.Sprintf("%s 小于等于 %s, 断言： 失败！", assertionText.Var, assertionText.Val)
						}
					} else {
						return AssertError, false, "不是数字类型，无法比较大小！"
					}
				case "float64":
					if value.(float64) <= num {
						return NoError, true, fmt.Sprintf("%s 小于等于 %s, 断言： 成功！", assertionText.Var, assertionText.Val)
					} else {
						return AssertError, false, fmt.Sprintf("%s 小于等于 %s, 断言： 失败！", assertionText.Var, assertionText.Val)
					}
				}

			} else {
				return AssertError, false, "不是数字类型，无法比较大小！"
			}
		default:
			return AssertError, false, "响应体断言条件不正确！"
		}
	}
	return AssertError, false, "未选择被断言体！"
}

type AssertionMsg struct {
	Type      string `json:"type"`
	Code      int64  `json:"code" bson:"code"`
	IsSucceed bool   `json:"isSucceed" bson:"isSucceed"`
	Msg       string `json:"msg" bson:"msg"`
}
