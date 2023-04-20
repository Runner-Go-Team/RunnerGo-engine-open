// Package tools  工具类

package tools

import (
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/tidwall/gjson"
	"os"
	"regexp"
	"strings"
	"time"
)

// TimeDifference 时间差，纳秒

func TimeDifference(startTime int64) (difference uint64) {
	endTime := time.Now().UnixNano()
	difference = uint64(endTime - startTime)
	return
}

func TimeDifference1(startTime int64) (difference uint64) {
	endTime := time.Now().UnixMilli()
	difference = uint64(endTime - startTime)
	return
}

// InArrayStr 判断字符串是否在数组内
func InArrayStr(str string, arr []string) (inArray bool) {
	for _, s := range arr {
		if s == str {
			inArray = true
			break
		}
	}
	return
}

// ToString 将map转化为string
func ToString(args map[string]interface{}) string {
	str, _ := json.Marshal(args)
	return string(str)
}

var SymbolList = []string{"`", "?", "~", "\\", "&", "*", "^", "%", "$", "￥", "#", "@", "!", "=", "+", "-", "_", "(", ")", "<", ">", ",", "."}
var PreFix = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
var LowerEnglishList = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
var UpperEnglishList = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "1", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}

// VariablesMatch 变量匹配, 如 name = {{name}}
func VariablesMatch(str string) (value string) {
	if strings.Contains(str, "{{") && strings.Contains(str, "}}") && strings.Index(str, "}}") > strings.Index(str, "{{") {
		value = FindDestStr(str, "{{(.*?)}}")
		for _, v := range PreFix {
			if strings.HasPrefix(value, v) {
				return str
			}
		}
		for _, v := range SymbolList {
			if strings.Contains(value, v) {
				return str
			}
		}
		return
	}

	return str
}

// FindDestStr 匹配规则
//func FindDestStr(str string, rex string) (result string) {
//	defer DeferPanic(fmt.Sprintf("正则表达式书写错误： %s", rex))
//	if strings.Contains(rex, "(.*?)") {
//		compileRegex := regexp.MustCompile(rex)
//		matchArr := compileRegex.FindStringSubmatch(str)
//		if len(matchArr) > 0 {
//			result = matchArr[len(matchArr)-1]
//		}
//		return
//	} else if strings.Contains(rex, "[0-9]+") {
//		compileRegex := regexp.MustCompile(rex)
//		matchArr := compileRegex.FindStringSubmatch(str)
//		if len(matchArr) > 0 {
//			result = matchArr[len(matchArr)-1]
//		}
//		rex = "[0-9]+"
//		compileRegex = regexp.MustCompile(rex)
//		matchArr = compileRegex.FindStringSubmatch(result)
//		if len(matchArr) > 0 {
//			result = matchArr[len(matchArr)-1]
//		}
//		return
//	}
//	return
//}

// FindDestStr 匹配规则
func FindDestStr(str string, rex string) (result string) {
	defer DeferPanic(fmt.Sprintf("正则表达式书写错误： %s", rex))
	compileRegex := regexp.MustCompile(rex)
	matchArr := compileRegex.FindStringSubmatch(str)
	if len(matchArr) > 0 {
		result = matchArr[len(matchArr)-1]
	}
	fmt.Println("result:::::  ", result)
	return
}

func DeferPanic(msg string) {
	if err := recover(); err != nil {
		log.Logger.Error(msg)
	}
}

// FindAllDestStr 匹配所有的
func FindAllDestStr(str, rex string) (result [][]string) {
	compileRegex := regexp.MustCompile(rex)
	result = compileRegex.FindAllStringSubmatch(str, -1)
	return
}

// PathExists 判断文件或文件夹是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			mkErr := os.MkdirAll(path, os.ModePerm)
			if mkErr != nil {
				log.Logger.Error(fmt.Sprintf("机器ip:%s, 创建文件夹失败", middlewares.LocalIp))
				return false
			}
		}
		return false
	}
	if os.IsNotExist(err) {
		mkErr := os.MkdirAll(path, os.ModePerm)
		if mkErr != nil {
			log.Logger.Error(fmt.Sprintf("机器ip:%s, 创建文件夹失败", middlewares.LocalIp))
			return false
		}
	}
	return true

}

// JsonPath json格式提取数据

func JsonPath(source, expression string) (value string) {
	gq := gjson.Get(source, expression)
	value = gq.String()
	return
}

func Decode(str string) (value string) {
	return
}
