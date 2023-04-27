package tools

import (
	"fmt"
	idvalidator "github.com/guanguans/id-validator"
	uuid "github.com/satori/go.uuid"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var ControllerMapsType = make(map[string]interface{})

// RandomFloat0 随机生成0-1之间的小数
func RandomFloat0() float64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Float64()
}

// RandomString 从list中随机生成n个字符组成的字符出
func RandomString(x string) (str string) {
	list := append(UpperEnglishList, LowerEnglishList...)
	list = append(list, PreFix...)
	rand.Seed(time.Now().UnixNano())
	n, err := strconv.Atoi(x)
	if err != nil {
		return
	}
	for i := 0; i < n; i++ {
		index := rand.Intn(len(list) - 0)
		str = str + list[index]
	}
	return
}

// RandomInt 生成min-max之间的随机数
func RandomInt(min, max string) string {
	n, err := strconv.Atoi(min)
	m, err := strconv.Atoi(max)
	if err != nil {
		return ""
	}
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%d", rand.Intn(m-n)+n)
}

// IdCard 根据参数生成身份证号
// isEighteen 是否生成18位号码
// address    省市县三级地区官方全称：如`北京市`、`台湾省`、`香港特别行政区`、`深圳市`、`黄浦区`
// birthday   出生日期：如 `2000`、`198801`、`19990101`
// sex        性别：1为男性，0为女性
func IdCard(isEighteen string, address string, birthday string, sex string) string {
	var isEigh bool
	var sexI int
	if isEighteen == "true" {
		isEigh = true
	}
	sexI, err := strconv.Atoi(sex)
	if err != nil {
		sexI = 1
	}
	return idvalidator.FakeRequireId(isEigh, address, birthday, sexI)
}

// RandomIdCard 随机生成身份证号
func RandomIdCard() string {
	return idvalidator.FakeId()
}

// VerifyIdCard 验证身份证号是否合法
func VerifyIdCard(cardId string, strict string) bool {
	if strict == "false" {
		return idvalidator.IsValid(cardId, false)
	}
	return idvalidator.IsValid(cardId, true)
}

// ToStringLU 改变字符串大小写
func ToStringLU(str, option string) string {
	if str == "" {
		return str
	}

	switch option {
	case "L":
		return strings.ToLower(str)
	default:
		return strings.ToUpper(str)
	}
}

func GetUUid() string {
	uid, _ := uuid.NewV4()
	return uid.String()
}

// ToTimeStamp 时间戳
func ToTimeStamp(option string) string {
	times := time.Now()
	switch option {
	case "s":
		return fmt.Sprintf("%d", times.Unix())
	case "ms":

		return fmt.Sprintf("%d", times.UnixMilli())

	case "ns":
		return fmt.Sprintf("%d", times.UnixNano())

	case "ws":
		return fmt.Sprintf("%d", times.UnixMicro())
	default:
		return fmt.Sprintf("%d", times.Unix())
	}
}

func InitPublicFunc() {
	ControllerMapsType["RandomFloat0"] = RandomFloat0
	ControllerMapsType["RandomString"] = RandomString
	ControllerMapsType["RandomInt"] = RandomInt
	ControllerMapsType["IdCard"] = IdCard
	ControllerMapsType["RandomIdCard"] = RandomIdCard
	ControllerMapsType["VerifyIdCard"] = VerifyIdCard
	ControllerMapsType["MD5"] = MD5
	ControllerMapsType["SHA256"] = SHA256
	ControllerMapsType["SHA512"] = SHA512
	ControllerMapsType["SHA1"] = SHA1
	ControllerMapsType["SHA224"] = SHA224
	ControllerMapsType["SHA384"] = SHA384
	ControllerMapsType["ToStringLU"] = ToStringLU
	ControllerMapsType["ToTimeStamp"] = ToTimeStamp
	ControllerMapsType["GetUUid"] = GetUUid
}

func CallPublicFunc(funcName string, parameters []string) []reflect.Value {

	if function, ok := ControllerMapsType[funcName]; ok {
		f := reflect.ValueOf(function)
		if len(parameters) != f.Type().NumIn() {
			return nil
		}

		in := make([]reflect.Value, len(parameters))
		if len(parameters) == 0 {
			return f.Call(in)
		}

		for k, param := range parameters {
			param = strings.TrimSpace(param)
			in[k] = reflect.ValueOf(param)
		}
		return f.Call(in)

	}
	return nil
}

const (
	StringTypeGo = "string"
	IntType      = "int"
	Float64Type  = "float64"
	BoolType     = "bool"
)

func ParsFunc(source string) (value string) {
	if !strings.HasSuffix(source, "__") || !strings.HasPrefix(source, "__") || !strings.Contains(source, "(") || !strings.Contains(source, ")") || strings.IndexAny(source, "(") > strings.IndexAny(source, ")") {
		value = source
		return
	}
	var parameters []string
	key := strings.Split(source, "__")[1]
	list := strings.Split(key, "(")
	funcName := list[0]
	if len(list) <= 1 {
		value = source
		return
	}
	if len(list[1]) > 1 {
		parameters = strings.Split(list[1], ")")
		parameters = parameters[0 : len(parameters)-1]
		if strings.Contains(parameters[0], ",") {
			parameters = strings.Split(parameters[0], ",")
		}
	}
	reflectValue := CallPublicFunc(funcName, parameters)
	if reflectValue == nil {
		value = source
		return
	}

	switch reflectValue[0].Type().String() {
	case StringTypeGo:
		value = reflectValue[0].String()
	case BoolType:
		if reflectValue[0].Bool() {
			value = "true"
		} else {
			value = "false"
		}
	case IntType:
		value = fmt.Sprintf("%d", reflectValue[0].Int())
	case Float64Type:
		value = fmt.Sprintf("%f", reflectValue[0].Float())
	}
	return
}
