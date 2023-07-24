package constant

const (
	NILSTRING = ""
	NILINT    = 0
)

// Form 支持协议类型
const (
	FormTypeHTTP      = "api"       // http协议
	FormTypeWebSocket = "websocket" // webSocket协议
	FormTypeDubbo     = "dubbo"     // grpc协议
	FormTypeTcp       = "tcp"
	FormTypeSql       = "sql"
	HTTP              = "http:"
	HTTPS             = "https:"
)

// 返回 code 码
const (
	// NoError 没有错误
	NoError = int64(10000)
	// AssertError 断言错误
	AssertError = int64(10001)
	// RequestError 请求错误
	RequestError = int64(10002)
	// ServiceError 服务错误
	ServiceError = int64(10003)
)

// Dubbo 值类型
const (
	JavaString    = "java.lang.String"
	JavaInteger   = "java.lang.Integer"
	JavaDouble    = "java.lang.Double"
	JavaShort     = "java.lang.Short"
	JavaLong      = "java.lang.Long"
	JavaFloat     = "java.lang.Float"
	JavaByte      = "java.lang.Byte"
	JavaBoolean   = "java.lang.Boolean"
	JavaCharacter = "java.lang.Character"
	JavaMap       = "java.lang.Map"
	JavaList      = "java.lang.List"
)

// 文本断言类型
const (
	ResponseHeaders = 1 // 断言响应的信息头
	ResponseData    = 2 // 断言响应的body信息
	ResponseCode    = 3 // 断言响应码
)

// 事件类型
const (
	RequestType        = "api"                  // 接口请求
	IfControllerType   = "condition_controller" // if控制器
	WaitControllerType = "wait_controller"      // 等待控制器
	SqlType            = "sql"                  // sql
	RedisType          = "redis"
	MqttType           = "mqtt"
	KafkaType          = "kafka"
)

// 逻辑运算符
const (
	Equal              = "eq"         // 等于
	UNEqual            = "uneq"       // 不等于
	GreaterThan        = "gt"         // 大于
	GreaterThanOrEqual = "gte"        // 大于或等于
	LessThan           = "lt"         // 小于
	LessThanOrEqual    = "lte"        // 小于或等于
	Includes           = "includes"   // 包含
	UNIncludes         = "unincludes" // 不包含
	NULL               = "null"       // 为空
	NotNULL            = "notnull"    // 不为空

	OriginatingFrom = "以...开始"
	EndIn           = "以...结束"
)

// 数据类型
const (
	StringType    = "String"
	TextType      = "Text"
	ObjectType    = "Object"
	ArrayType     = "Array"
	IntegerType   = "Integer"
	NumberType    = "Number"
	FloatType     = "Float"
	DoubleType    = "Double"
	FileType      = "File"
	FileUrlType   = "FileUrl"
	DateType      = "Date"
	DateTimeType  = "DateTime"
	TimeStampType = "TimeStampType"
	BooleanType   = "boolean"
	InterfaceMap  = "map[interface {}]interface {}"
)

// body 格式
const (
	NoneMode      = "none"
	FormMode      = "form-data"
	UrlencodeMode = "urlencoded"
	JsonMode      = "json"
	XmlMode       = "xml"
	JSMode        = "javascript"
	PlainMode     = "plain"
	HtmlMode      = "html"
	MysqlMode     = "sql"
)

// 时间运行状态
const (
	Success = "success" // 成功
	Failed  = "failed"  // 失败
	End     = "end"     // 结束
	NotHit  = "not_hit" // 未命中
	NotRun  = "not_run" // 未运行
)

// debug日志状态
const (
	All         = "all"
	OnlyError   = "only_error"
	OnlySuccess = "only_success"
	StopDebug   = "stop"
)

// 关联提取类型
const (
	RegExtract    = 0
	JsonExtract   = 1
	HeaderExtract = 2
	CodeExtract   = 3
)

// 开关
const (
	Open  = 1
	Close = 2
)

// 运行类型
const (
	SceneType = "scene"
	PlanType  = "plan"
)

// 认证类型
const (
	NoAuth         = "noauth"
	Unidirectional = "unidirectional"
	Bidirectional  = "bidirectional"
	Kv             = "kv"
	BEarer         = "bearer"
	BAsic          = "basic"
	DigestType     = "digest"
	HawkType       = "hawk"
	Awsv4Type      = "awsv4"
	NtlmType       = "ntlm"
	EdgegridType   = "edgegrid"
	Oauth1Type     = "oauth1"
)

const (
	CentralizedMode = 0
	AloneMode       = 1
)

const (
	AuToOrderMode = int64(1)
	AuToSameMode  = int64(2)
)

// 1: stopPlan; 2: debug; 3.报告变更
const (
	StopPlan     = 1
	DebugStatus  = 2
	ReportChange = 3
)

// tcp/ws 消息类型
const (
	MsBinary           = "Binary"
	MsText             = "Text"
	MsJson             = "Json"
	MsXml              = "Xml"
	LongConnection     = int32(1)
	ShortConnection    = int32(2)
	AutoConnectionSend = int32(1)
	ConnectionAndSend  = int32(0)
	UnConnection       = int32(1)
	SendMessage        = int32(2)
)
