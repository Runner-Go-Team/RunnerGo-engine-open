package config

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strconv"
	"time"
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
	Name                     string        `yaml:"name"`
	Address                  string        `yaml:"address"`
	Version                  string        `yaml:"version"`
	ReadTimeout              time.Duration `yaml:"readTimeout"`
	WriteTimeout             time.Duration `yaml:"writeTimeout"`
	MaxConnPerHost           int           `yaml:"maxConnPerHost"`
	MaxIdleConnDuration      time.Duration
	MaxConnWaitTimeout       time.Duration
	NoDefaultUserAgentHeader bool `yaml:"noDefaultUserAgentHeader"`
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
	NotifyStopStress = "https://127.0.0.1:8080/management/api/v1/plan/notify_stop_stress"
	NotifyRunFinish  = "https://127.0.0.1:8080/management/api/v1/plan/notify_run_finish"
	Region           = "北京"
	Port             = 8002
	MongoData        = "runnergo_open"
	StressDebug      = "stress_debug"
	SceneDebugTable  = "scene_debug"
	ApiDebugTable    = "api_debug"
	DebugStatusTable = "debug_status"
	AutoTable        = "auto_report"
	RedisAddress     = "127.0.0.1:6379"
	KafkaTopic       = "report"
	KafkaAddress     = "127.0.0.1:27017"
	HttpAddress      = "0.0.0.0:8002"
)

func initLog() {
	logPath := os.Getenv("RUNNER_GO_ENGINE_LOG_PATH")
	if logPath == "" {
		logPath = LogPath
	}
	Conf.Log.Path = logPath
}
func initManagement() {
	var management Management
	notifyStopStress := os.Getenv("RUNNER_GO_MANAGEMENT_NOTIFY_STOP_STRESS")
	if notifyStopStress == "" {
		notifyStopStress = NotifyStopStress
	}
	management.NotifyStopStress = notifyStopStress
	notifyRunFinish := os.Getenv("RUNNER_GO_MANAGEMENT_NOTIFY_RUN_FINISH")
	if notifyRunFinish == "" {
		notifyRunFinish = NotifyRunFinish
	}
	management.NotifyRunFinish = notifyRunFinish
	Conf.Management = management
}

func initMachine() {
	var runnerGoMachine Machine
	maxGo, err := strconv.Atoi(os.Getenv("RUNNER_GO_MACHINE_MAX_GOROUTINES"))
	if err != nil {
		maxGo = 20000
	}
	runnerGoMachine.MaxGoroutines = maxGo
	serverType, err := strconv.Atoi(os.Getenv("RUNNER_GO_MACHINE_SERVER_TYPE"))
	if err != nil {
		serverType = 0
	}
	runnerGoMachine.ServerType = serverType
	runnerGoMachine.NetName = os.Getenv("RUNNER_GO_MACHINE_NET_NAME")
	runnerGoMachine.DiskName = os.Getenv("RUNNER_GO_MACHINE_DISK_NAME")
	Conf.Machine = runnerGoMachine
}

func initHeartbeat() {
	var runnerGoHeartbeat Heartbeat
	port, err := strconv.Atoi(os.Getenv("RUNNER_GO_HEARTBEAT_PORT"))
	if err != nil {
		port = Port
	}
	runnerGoHeartbeat.Port = int32(port)
	region := os.Getenv("RUNNER_GO_HEARTBEAT_REGION")
	if region == "" {
		region = Region
	}
	runnerGoHeartbeat.Region = region
	duration, err := strconv.ParseInt(os.Getenv("RUNNER_GO_HEARTBEAT_DURATION"), 10, 64)
	if err != nil {
		duration = 3
	}
	runnerGoHeartbeat.Duration = duration
	resources, err := strconv.ParseInt(os.Getenv("RUNNER_GO_HEARTBEAT_RESOURCES"), 10, 64)
	if err != nil {
		resources = 3
	}
	runnerGoHeartbeat.Resources = resources
	Conf.Heartbeat = runnerGoHeartbeat
}

// 初始化mongo
func initMongo() {
	var runnerGoMongo Mongo
	runnerGoMongo.Password = os.Getenv("RUNNER_GO_MONGO_PASSWORD")
	dsn := os.Getenv("RUNNER_GO_MONGO_DSN")
	if dsn == "" {
		dsn = fmt.Sprintf("mongodb://runnergo_open:%s@127.0.0.1:27017/runnergo_open", runnerGoMongo.Password)
	}
	runnerGoMongo.DSN = dsn
	runnerGoMongo.DataBase = MongoData
	runnerGoMongo.StressDebugTable = StressDebug
	runnerGoMongo.DebugStatusTable = DebugStatusTable
	runnerGoMongo.SceneDebugTable = SceneDebugTable
	runnerGoMongo.ApiDebugTable = ApiDebugTable
	runnerGoMongo.AutoTable = AutoTable
	Conf.Mongo = runnerGoMongo
}

func initRedis() {
	var runnerGoRedis Redis
	address := os.Getenv("RUNNER_GO_REDIS_ADDRESS")
	if address == "" {
		address = RedisAddress
	}
	runnerGoRedis.Address = address
	runnerGoRedis.Password = os.Getenv("RUNNER_GO_REDIS_PASSWORD")
	db, err := strconv.ParseInt(os.Getenv("RUNNER_GO_DB"), 10, 64)
	if err != nil {
		db = 0
	}
	runnerGoRedis.DB = db
	Conf.Redis = runnerGoRedis
}

func initKafka() {
	var runnerGoKafka Kafka
	topic := os.Getenv("RUNNER_GO_KAFKA_TOPIC")
	if topic == "" {
		topic = KafkaTopic
	}
	runnerGoKafka.TopIc = topic
	address := os.Getenv("RUNNER_GO_KAFKA_ADDRESS")
	if address == "" {
		address = KafkaAddress
	}
	runnerGoKafka.Address = address
	Conf.Kafka = runnerGoKafka

}

func initHttp() {
	var http Http
	http.Name = os.Getenv("ENGINE_HTTP_NAME")
	address := os.Getenv("ENGINE_HTTP_ADDRESS")
	if address == "" {
		address = HttpAddress
	}
	http.Address = address
	http.Version = os.Getenv("HTTP_VERSION")
	readTimeout, err := strconv.ParseInt(os.Getenv("HTTP_READ_TIMEOUT"), 10, 64)
	if err != nil {
		readTimeout = 0
	}
	http.ReadTimeout = time.Duration(readTimeout)

	writeTimeout, err := strconv.ParseInt(os.Getenv("HTTP_WRITE_TIMEOUT"), 10, 64)
	if err != nil {
		writeTimeout = 0
	}
	http.WriteTimeout = time.Duration(writeTimeout)

	maxConnPerHost, err := strconv.Atoi(os.Getenv("HTTP_MAX_CONN_PER_HOST"))
	if err != nil {
		maxConnPerHost = 0
	}
	http.MaxConnPerHost = maxConnPerHost

	httpMaxIdleConnDuration, err := strconv.ParseInt(os.Getenv("HTTP_MAX_IDLE_CONN_DURATION"), 10, 64)
	if err != nil {
		httpMaxIdleConnDuration = 0
	}
	http.MaxIdleConnDuration = time.Duration(httpMaxIdleConnDuration)

	httpMaxConnWaitTimeout, err := strconv.ParseInt(os.Getenv("HTTP_MAX_CONN_WAIT_TIMEOUT"), 10, 64)
	if err != nil {
		httpMaxConnWaitTimeout = 0
	}
	http.MaxConnWaitTimeout = time.Duration(httpMaxConnWaitTimeout)
	if os.Getenv("HTTP_NO_DEFAULT_USER_AGENT_HEADER") == "true" {
		http.NoDefaultUserAgentHeader = true
	} else {
		http.NoDefaultUserAgentHeader = false
	}
	Conf.Http = http
}
