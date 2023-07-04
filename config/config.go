package config

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strconv"
)

var Conf Config

type Config struct {
	Http        Http        `yaml:"http"`
	Kafka       Kafka       `yaml:"kafka"`
	ReportRedis ReportRedis `yaml:"reportRedis"`
	Redis       Redis       `yaml:"redis"`
	Mongo       Mongo       `yaml:"mongo"`
	Heartbeat   Heartbeat   `yaml:"heartbeat"`
	Machine     Machine     `yaml:"machine"`
	Management  Management  `yaml:"management"`
	Log         Log         `yaml:"log"`
}

type Log struct {
	Path string `yaml:"path"`
}
type Management struct {
	NotifyStopStress string `yaml:"notifyStopStress"`
	NotifyRunFinish  string `yaml:"notifyRunFinish"`
}

type Machine struct {
	MaxGoroutines int    `yaml:"maxGoroutines"`
	ServerType    int    `yaml:"serverType"`
	NetName       string `yaml:"netName"`
	DiskName      string `yaml:"diskName"`
}

type Heartbeat struct {
	Port      int32  `yaml:"port"`
	Region    string `yaml:"region"`
	Duration  int64  `yaml:"duration"`
	Resources int64  `yaml:"resources"`
}
type Http struct {
	Address string `yaml:"address"`
}

type Kafka struct {
	Address string `yaml:"address"`
	TopIc   string `yaml:"topIc"`
}

type ReportRedis struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int64  `yaml:"DB"`
}
type Redis struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int64  `yaml:"DB"`
}

type Mongo struct {
	DSN              string `yaml:"dsn"`
	Password         string `yaml:"password"`
	DataBase         string `yaml:"database"`
	StressDebugTable string `yaml:"stressDebugTable"`
	DebugStatusTable string `yaml:"debugTable"`
	SceneDebugTable  string `yaml:"sceneDebugTable"`
	ApiDebugTable    string `yaml:"apiDebugTable"`
	AutoTable        string `yaml:"autoTable"`
	SqlDebugTable    string `yaml:"sqlDebugTable"`
	TcpDebugTable    string `yaml:"tcpDebugTable"`
	WsDebugTable     string `yaml:"wsDebugTable"`
	DubboDebugTable  string `yaml:"dubboDebugTable"`
	MqttDebugTable   string `yaml:"mqttDebugTable"`
}

func InitConfig() {

	var conf string
	flag.StringVar(&conf, "c", "./open.yaml", "配置文件,默认为conf文件夹下的open文件")
	if !flag.Parsed() {
		flag.Parse()
	}

	viper.SetConfigFile(conf)
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	if err = viper.Unmarshal(&Conf); err != nil {
		panic(fmt.Errorf("unmarshal error config file: %w", err))
	}

	fmt.Println("config initialized")

}

// EnvInitConfig 读取环境变量
func EnvInitConfig() {
	initLog()
	initManagement()
	initMachine()
	initHeartbeat()
	initMongo()
	initRedis()
	initKafka()
	initHttp()
}

const (
	LogPath          = "/data/logs/RunnerGo/RunnerGo-engine-info.log"
	NotifyStopStress = "https://127.0.0.0:30000/management/api/v1/plan/notify_stop_stress"
	NotifyRunFinish  = "https://127.0.0.0:30000/management/api/v1/auto_plan/notify_run_finish"
	Region           = "北京"
	Port             = 30000
	MongoData        = "runnergo"
	StressDebug      = "stress_debug"
	SceneDebugTable  = "scene_debug"
	ApiDebugTable    = "api_debug"
	DebugStatusTable = "debug_status"
	AutoTable        = "auto_report"
	SqlDebugTable    = "sql_debug"
	TcpDebugTable    = "tcp_debug"
	RedisAddress     = "127.0.0.0:6379"
	KafkaTopic       = "report"
	KafkaAddress     = "127.0.0.0:9092"
	HttpAddress      = "0.0.0.0:30000"
)

func initLog() {
	logPath := os.Getenv("RG_ENGINE_LOG_PATH")
	if logPath == "" {
		logPath = LogPath
	}
	Conf.Log.Path = logPath
}
func initManagement() {
	var management Management
	notifyStopStress := os.Getenv("RG_MANAGEMENT_NOTIFY_STOP_STRESS")
	if notifyStopStress == "" {
		notifyStopStress = NotifyStopStress
	}
	management.NotifyStopStress = notifyStopStress
	notifyRunFinish := os.Getenv("RG_MANAGEMENT_NOTIFY_RUN_FINISH")
	if notifyRunFinish == "" {
		notifyRunFinish = NotifyRunFinish
	}
	management.NotifyRunFinish = notifyRunFinish
	Conf.Management = management
}

func initMachine() {
	var runnerGoMachine Machine
	maxGo, err := strconv.Atoi(os.Getenv("RG_MACHINE_MAX_GOROUTINES"))
	if err != nil {
		maxGo = 20000
	}
	runnerGoMachine.MaxGoroutines = maxGo
	serverType, err := strconv.Atoi(os.Getenv("RG_MACHINE_SERVER_TYPE"))
	if err != nil {
		serverType = 0
	}
	runnerGoMachine.ServerType = serverType
	runnerGoMachine.NetName = os.Getenv("RG_MACHINE_NET_NAME")
	runnerGoMachine.DiskName = os.Getenv("RG_MACHINE_DISK_NAME")
	Conf.Machine = runnerGoMachine
}

func initHeartbeat() {
	var runnerGoHeartbeat Heartbeat
	port, err := strconv.Atoi(os.Getenv("RG_HEARTBEAT_PORT"))
	if err != nil {
		port = Port
	}
	runnerGoHeartbeat.Port = int32(port)
	region := os.Getenv("RG_HEARTBEAT_REGION")
	if region == "" {
		region = Region
	}
	runnerGoHeartbeat.Region = region
	duration, err := strconv.ParseInt(os.Getenv("RG_HEARTBEAT_DURATION"), 10, 64)
	if err != nil {
		duration = 3
	}
	runnerGoHeartbeat.Duration = duration
	resources, err := strconv.ParseInt(os.Getenv("RG_HEARTBEAT_RESOURCES"), 10, 64)
	if err != nil {
		resources = 3
	}
	runnerGoHeartbeat.Resources = resources
	Conf.Heartbeat = runnerGoHeartbeat
}

// 初始化mongo
func initMongo() {
	var runnerGoMongo Mongo
	runnerGoMongo.Password = os.Getenv("RG_MONGO_PASSWORD")
	dsn := os.Getenv("RG_MONGO_DSN")
	if dsn == "" {
		dsn = fmt.Sprintf("mongodb://runnergo:%s@127.0.0.0:27017/runnergo", runnerGoMongo.Password)
	}
	runnerGoMongo.DSN = dsn
	mongoData := os.Getenv("RG_MONGO_DATABASE")
	if mongoData == "" {
		runnerGoMongo.DataBase = MongoData
	} else {
		runnerGoMongo.DataBase = mongoData
	}
	stressDebug := os.Getenv("RG_MONGO_STRESS_DEBUG_TABLE")
	if stressDebug == "" {
		runnerGoMongo.StressDebugTable = StressDebug
	} else {
		runnerGoMongo.StressDebugTable = stressDebug
	}
	debugStatusTable := os.Getenv("RG_MONGO_DEBUG_TABLE")
	if debugStatusTable == "" {
		runnerGoMongo.DebugStatusTable = DebugStatusTable
	} else {
		runnerGoMongo.DebugStatusTable = debugStatusTable
	}
	sceneDebugTable := os.Getenv("RG_MONGO_SCENE_DEBUG_TABLE")
	if stressDebug == "" {
		runnerGoMongo.SceneDebugTable = SceneDebugTable
	} else {
		runnerGoMongo.SceneDebugTable = sceneDebugTable
	}
	apiDebugTable := os.Getenv("RG_MONGO_API_DEBUG_TABLE")
	if apiDebugTable == "" {
		runnerGoMongo.ApiDebugTable = ApiDebugTable
	} else {
		runnerGoMongo.ApiDebugTable = apiDebugTable
	}
	autoTable := os.Getenv("RG_MONGO_AUTO_TABLE")
	if autoTable == "" {
		runnerGoMongo.AutoTable = AutoTable
	} else {
		runnerGoMongo.AutoTable = autoTable
	}
	sqlDebugTable := os.Getenv("RG_MONGO_SQL_DEBUG_TABLE")
	if sqlDebugTable == "" {
		runnerGoMongo.SqlDebugTable = SqlDebugTable
	} else {
		runnerGoMongo.SqlDebugTable = sqlDebugTable
	}
	tcpDebugTable := os.Getenv("RG_MONGO_TCP_DEBUG_TABLE")
	if tcpDebugTable == "" {
		runnerGoMongo.TcpDebugTable = TcpDebugTable
	} else {
		runnerGoMongo.TcpDebugTable = tcpDebugTable
	}
	Conf.Mongo = runnerGoMongo
}

func initRedis() {
	var runnerGoRedis Redis
	var reportRedis ReportRedis
	address := os.Getenv("RG_REDIS_ADDRESS")
	if address == "" {
		address = RedisAddress
	}
	reportRedis.Address = address
	runnerGoRedis.Address = reportRedis.Address
	runnerGoRedis.Password = os.Getenv("RG_REDIS_PASSWORD")
	reportRedis.Password = runnerGoRedis.Password
	db, err := strconv.ParseInt(os.Getenv("RG_REDIS_DB"), 10, 64)
	if err != nil {
		db = 0
	}
	reportRedis.DB = db
	runnerGoRedis.DB = reportRedis.DB

	Conf.Redis = runnerGoRedis
	Conf.ReportRedis = reportRedis

}

func initKafka() {
	var runnerGoKafka Kafka
	topic := os.Getenv("RG_KAFKA_TOPIC")
	if topic == "" {
		topic = KafkaTopic
	}
	runnerGoKafka.TopIc = topic
	address := os.Getenv("RG_KAFKA_ADDRESS")
	if address == "" {
		address = KafkaAddress
	}
	runnerGoKafka.Address = address
	Conf.Kafka = runnerGoKafka

}

func initHttp() {
	var http Http
	address := os.Getenv("RG_ENGINE_HTTP_ADDRESS")
	if address == "" {
		address = HttpAddress
	}
	http.Address = address

	Conf.Http = http
}
