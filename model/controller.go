package model

import (
	"strings"
)

func (ic *Event) PerForm(value string) (result, msg string) {
	switch ic.Compare {
	case Equal:
		if strings.Compare(ic.Val, value) == 0 {
			result = Success
			msg = ic.Val + " " + Equal + " " + value + "  成功"
		} else {
			result = Failed
			msg = ic.Val + " " + Equal + " " + value + "  失败"
		}

	case UNEqual:
		if strings.Compare(ic.Val, value) != 0 {
			result = Success
			msg = ic.Val + " " + UNEqual + " " + value + "  成功"
		} else {
			result = Failed
			msg = ic.Val + " " + UNEqual + " " + value + "  失败"
		}

	case GreaterThan:
		if value > ic.Val {
			result = Success
			msg = ic.Val + " " + GreaterThan + " " + value + "  成功"
		} else {
			result = Failed
			msg = ic.Val + " " + GreaterThan + " " + value + "  失败"
		}

	case GreaterThanOrEqual:
		if value >= ic.Val {
			result = Success
			msg = ic.Val + " " + GreaterThanOrEqual + " " + value + "  成功"
		} else {
			result = Failed
			msg = ic.Val + " " + GreaterThanOrEqual + " " + value + "  失败"
		}

	case LessThan:
		if value < ic.Val {
			result = Success
			msg = ic.Val + " " + LessThan + " " + value + "  成功"
		} else {
			result = Failed
			msg = ic.Val + " " + LessThan + " " + value + "  失败"
		}

	case LessThanOrEqual:
		if value <= ic.Val {
			result = Success
			msg = ic.Val + " " + LessThanOrEqual + " " + value + "  成功"
		} else {
			result = Failed
			msg = ic.Val + " " + LessThanOrEqual + " " + value + "  失败"
		}
	case Includes:
		if strings.Contains(value, ic.Val) {
			result = Success
			msg = ic.Val + " " + Includes + " " + value + "  成功"
		} else {
			result = Failed
			msg = ic.Val + " " + Includes + " " + value + "  失败"
		}
	case UNIncludes:
		if !strings.Contains(value, ic.Val) {
			result = Success
			msg = ic.Val + " " + UNIncludes + " " + value + "  成功"
		} else {
			result = Failed
			msg = ic.Val + " " + UNIncludes + " " + value + "  失败"
		}
	case NULL:
		if value == "" {
			result = Success
			msg = ic.Val + " " + NULL + " " + value + "  成功"
		} else {
			result = Failed
			msg = ic.Val + " " + NULL + " " + value + "  失败"
		}

	case NotNULL:
		if value != "" {
			result = Success
			msg = ic.Val + " " + NotNULL + " " + value + "  成功"
		} else {
			result = Failed
			msg = ic.Val + " " + NotNULL + " " + value + "  失败"
		}
	default:
		result = Failed
		msg = ""
	}
	return
}
