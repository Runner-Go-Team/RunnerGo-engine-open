package model

import (
	"github.com/Runner-Go-Team/RunnerGo-engine-open/constant"
	"strings"
)

func (ic *Event) PerForm(value string) (result string, msg string) {
	switch ic.Compare {
	case constant.Equal:
		if strings.Compare(ic.Val, value) == 0 {
			result = constant.Success
			msg = ic.Val + " " + constant.Equal + " " + value + "  成功"
		} else {
			result = constant.Failed
			msg = ic.Val + " " + constant.Equal + " " + value + "  失败"
		}

	case constant.UNEqual:
		if strings.Compare(ic.Val, value) != 0 {
			result = constant.Success
			msg = ic.Val + " " + constant.UNEqual + " " + value + "  成功"
		} else {
			result = constant.Failed
			msg = ic.Val + " " + constant.UNEqual + " " + value + "  失败"
		}

	case constant.GreaterThan:
		if value > ic.Val {
			result = constant.Success
			msg = ic.Val + " " + constant.GreaterThan + " " + value + "  成功"
		} else {
			result = constant.Failed
			msg = ic.Val + " " + constant.GreaterThan + " " + value + "  失败"
		}

	case constant.GreaterThanOrEqual:
		if value >= ic.Val {
			result = constant.Success
			msg = ic.Val + " " + constant.GreaterThanOrEqual + " " + value + "  成功"
		} else {
			result = constant.Failed
			msg = ic.Val + " " + constant.GreaterThanOrEqual + " " + value + "  失败"
		}

	case constant.LessThan:
		if value < ic.Val {
			result = constant.Success
			msg = ic.Val + " " + constant.LessThan + " " + value + "  成功"
		} else {
			result = constant.Failed
			msg = ic.Val + " " + constant.LessThan + " " + value + "  失败"
		}

	case constant.LessThanOrEqual:
		if value <= ic.Val {
			result = constant.Success
			msg = ic.Val + " " + constant.LessThanOrEqual + " " + value + "  成功"
		} else {
			result = constant.Failed
			msg = ic.Val + " " + constant.LessThanOrEqual + " " + value + "  失败"
		}
	case constant.Includes:
		if strings.Contains(value, ic.Val) {
			result = constant.Success
			msg = ic.Val + " " + constant.Includes + " " + value + "  成功"
		} else {
			result = constant.Failed
			msg = ic.Val + " " + constant.Includes + " " + value + "  失败"
		}
	case constant.UNIncludes:
		if !strings.Contains(value, ic.Val) {
			result = constant.Success
			msg = ic.Val + " " + constant.UNIncludes + " " + value + "  成功"
		} else {
			result = constant.Failed
			msg = ic.Val + " " + constant.UNIncludes + " " + value + "  失败"
		}
	case constant.NULL:
		if value == "" {
			result = constant.Success
			msg = ic.Val + " " + constant.NULL + " " + value + "  成功"
		} else {
			result = constant.Failed
			msg = ic.Val + " " + constant.NULL + " " + value + "  失败"
		}

	case constant.NotNULL:
		if value != "" {
			result = constant.Success
			msg = ic.Var + " " + constant.NotNULL + " " + value + "  成功"
		} else {
			result = constant.Failed
			msg = ic.Var + " " + constant.NotNULL + " " + value + "  失败"
		}
	default:
		result = constant.Failed
		msg = ""
	}
	return
}
