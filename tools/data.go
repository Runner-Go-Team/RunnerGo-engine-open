package tools

import (
	"encoding/base64"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"strings"
)

// Base64DeEncode base64解码
func Base64DeEncode(str string, dataType string) (decoded []byte, fileType string) {
	if dataType != "File" {
		return
	}
	strs := strings.Split(str, ";base64,")
	if len(strs) < 2 {
		return
	}
	str = strs[1]
	strType := strings.Split(strs[0], "data:")
	if len(strType) < 2 {
		return
	}
	fileType = strType[1]

	if str[len(str)-1] == 61 {
		decoded, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			log.Logger.Error(fmt.Sprintf("机器ip:%s,  base64解码错误：%s", middlewares.LocalIp, err.Error()))
		}
		return decoded, fileType
	} else {
		decoded, err := base64.RawStdEncoding.DecodeString(str)
		if err != nil {
			log.Logger.Error(fmt.Sprintf("机器ip:%s, base64解码错误：%s", middlewares.LocalIp, err.Error()))
		}
		return decoded, fileType
	}
	return
}

// Base64Encode base64编码 常规编码，末尾不补 =
func Base64Encode(str string) (encode string) {
	msg := []byte(str)
	encode = base64.RawStdEncoding.EncodeToString(msg)
	return
}

// Base64EncodeStd 常规编码
func Base64EncodeStd(str string) string {
	msg := []byte(str)
	return base64.StdEncoding.EncodeToString(msg)
}
