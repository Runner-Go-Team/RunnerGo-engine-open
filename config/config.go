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
	Conf.Log.Path = os.Getenv("Runner_Go_Log_Path")
}
func initManagement() {
	var management Management
	management.Address = os.Getenv("Runner_Go_Management_Address")
	management.NotifyRunFinish = os.Getenv("Runner_Go_Management_NotifyRunFinish")
	Conf.Management = management
}

func initMachine() {
	var runnerGoMachine Machine
	maxGo, err := strconv.Atoi(os.Getenv("Runner_Go_Machine_Max_Goroutines"))
	if err != nil {
		maxGo = 20000
	}
	runnerGoMachine.MaxGoroutines = maxGo
	serverType, err := strconv.Atoi(os.Getenv("Runner_Go_Machine_Server_Type"))
	if err != nil {
		serverType = 0
	}
	runnerGoMachine.ServerType = serverType
	runnerGoMachine.NetName = os.Getenv("Runner_Go_Machine_Net_Name")
	runnerGoMachine.DiskName = os.Getenv("Runner_Go_Machine_Disk_Name")
	Conf.Machine = runnerGoMachine
}

func initHeartbeat() {
	var runnerGoHeartbeat Heartbeat
	port, err := strconv.Atoi(os.Getenv("Runner_Go_HeartBeat_Port"))
	if err != nil {
		port = 0
	}
	runnerGoHeartbeat.Port = int32(port)
	runnerGoHeartbeat.Region = os.Getenv("Runner_Go_HeartBeat_Region")
	duration, err := strconv.ParseInt(os.Getenv("Runner_Go_HeartBeat_Duration"), 10, 64)
	if err != nil {
		duration = 3
	}
	runnerGoHeartbeat.Duration = duration
	resources, err := strconv.ParseInt(os.Getenv("Runner_Go_HeartBeat_Resources"), 10, 64)
	if err != nil {
		resources = 3
	}
	runnerGoHeartbeat.Resources = resources
	Conf.Heartbeat = runnerGoHeartbeat
}

// 初始化mongo
func initMongo() {
	var runnerGoMongo Mongo
	runnerGoMongo.DSN = os.Getenv("Runner_Go_Mongo_DSN")
	runnerGoMongo.DataBase = os.Getenv("Runner_Go_Mongo_Database")
	runnerGoMongo.StressDebugTable = os.Getenv("Runner_Go_Mongo_Stress_Debug_Table")
	runnerGoMongo.DebugTable = os.Getenv("Runner_Go_Mongo_Debug_Table")
	runnerGoMongo.SceneDebugTable = os.Getenv("Runner_Go_Mongo_Scene_Debug_Table")
	runnerGoMongo.ApiDebugTable = os.Getenv("Runner_Go_Mongo_Api_Debug_Table")
	runnerGoMongo.AutoTable = os.Getenv("Runner_Go_Mongo_AutoTable")
	Conf.Mongo = runnerGoMongo
}

func initRedis() {
	var runnerGoRedis Redis
	runnerGoRedis.Address = os.Getenv("Runner_Go_Redis")
	runnerGoRedis.Password = os.Getenv("Runner_Go_Redis_Password")
	db, err := strconv.ParseInt(os.Getenv("Runner_Go_DB"), 10, 64)
	if err != nil {
		db = 0
	}
	runnerGoRedis.DB = db
	Conf.Redis = runnerGoRedis
}

func initKafka() {
	var runnerGokafka Kafka
	runnerGokafka.TopIc = os.Getenv("Runner_Go_Kafka_Topic")
	runnerGokafka.Address = os.Getenv("Runner_Go_Kafka_Address")
	Conf.Kafka = runnerGokafka

}

func initHttp() {
	var http Http
	http.Name = os.Getenv("Http_Name")
	http.Address = os.Getenv("Http_Address")
	http.Version = os.Getenv("Http_Version")
	readTimeout, err := strconv.ParseInt(os.Getenv("Http_Read_Timeout"), 10, 64)
	if err != nil {
		readTimeout = 0
	}
	http.ReadTimeout = time.Duration(readTimeout)

	writeTimeout, err := strconv.ParseInt(os.Getenv("Http_Write_Timeout"), 10, 64)
	if err != nil {
		writeTimeout = 0
	}
	http.WriteTimeout = time.Duration(writeTimeout)

	maxConnPerHost, err := strconv.Atoi(os.Getenv("Http_Max_Conn_Per_Host"))
	if err != nil {
		maxConnPerHost = 0
	}
	http.MaxConnPerHost = maxConnPerHost

	httpMaxIdleConnDuration, err := strconv.ParseInt(os.Getenv("Http_Max_Idle_Conn_Duration"), 10, 64)
	if err != nil {
		httpMaxIdleConnDuration = 0
	}
	http.MaxIdleConnDuration = time.Duration(httpMaxIdleConnDuration)

	httpMaxConnWaitTimeout, err := strconv.ParseInt(os.Getenv("Http_Max_Conn_Wait_Timeout"), 10, 64)
	if err != nil {
		httpMaxConnWaitTimeout = 0
	}
	http.MaxConnWaitTimeout = time.Duration(httpMaxConnWaitTimeout)
	if os.Getenv("Http_No_Default_User_Agent_Header") == "true" {
		http.NoDefaultUserAgentHeader = true
	} else {
		http.NoDefaultUserAgentHeader = false
	}
	Conf.Http = http
}
