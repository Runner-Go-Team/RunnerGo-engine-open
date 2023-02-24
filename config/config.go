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
	Address         string `yaml:"address"`
	NotifyRunFinish string `yaml:"notifyRunFinish"`
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
	DataBase         string `yaml:"database"`
	StressDebugTable string `yaml:"stressDebugTable"`
	DebugTable       string `yaml:"debugTable"`
	SceneDebugTable  string `yaml:"sceneDebugTable"`
	ApiDebugTable    string `yaml:"apiDebugTable"`
	AutoTable        string `yaml:"autoTable"`
}

func InitConfig() {

	var conf string
	flag.StringVar(&conf, "c", "./dev.yaml", "配置文件,默认为conf文件夹下的dev文件")
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

func initLog() {
	Conf.Log.Path = os.Getenv("RUNNER_GO_ENGINE_LOG_PATH")
}
func initManagement() {
	var management Management
	management.Address = os.Getenv("RUNNER_GO_MANAGEMENT_ADDRESS")
	management.NotifyRunFinish = os.Getenv("RUNNER_GO_MANAGEMENT_NOTIFY_RUN_FINISH")
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
		port = 0
	}
	runnerGoHeartbeat.Port = int32(port)
	runnerGoHeartbeat.Region = os.Getenv("RUNNER_GO_HEARTBEAT_REGION")
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
	runnerGoMongo.DSN = os.Getenv("RUNNER_GO_MONGO_DSN")
	runnerGoMongo.DataBase = os.Getenv("RUNNER_GO_MONGO_DATABASE")
	runnerGoMongo.StressDebugTable = os.Getenv("RUNNER_GO_MONGO_STRESS_DEBUG_TABLE")
	runnerGoMongo.DebugTable = os.Getenv("RUNNER_GO_MONGO_DEBUG_TABLE")
	runnerGoMongo.SceneDebugTable = os.Getenv("RUNNER_GO_MONGO_SCENE_DEBUG_TABLE")
	runnerGoMongo.ApiDebugTable = os.Getenv("RUNNER_GO_MONGO_API_DEBUG_TABLE")
	runnerGoMongo.AutoTable = os.Getenv("RUNNER_GO_MONGO_AUTO_TABLE")
	Conf.Mongo = runnerGoMongo
}

func initRedis() {
	var runnerGoRedis Redis
	runnerGoRedis.Address = os.Getenv("RUNNER_GO_REDIS_ADDRESS")
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
	runnerGoKafka.TopIc = os.Getenv("RUNNER_GO_KAFKA_TOPIC")
	runnerGoKafka.Address = os.Getenv("RUNNER_GO_KAFKA_ADDRESS")
	Conf.Kafka = runnerGoKafka

}

func initHttp() {
	var http Http
	http.Name = os.Getenv("ENGINE_HTTP_NAME")
	http.Address = os.Getenv("ENGINE_HTTP_ADDRESS")
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
