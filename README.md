# RunnerGo-collector-open

```text
engine 服务为发压服务，主要是发送请求
```
## 本地部署

# 下载，使用git命令下载到本地
```gitignore
 git clone https://github.com/Runner-Go-Team/RunnerGo-engine-open.git
```

# 修改配置文件， open.yaml
```yaml
heartbeat:
  port: 8002
  region: "北京"
  duration: 2
  resources: 5


http:
  address: "0.0.0.0:8002"                                    #本服务host
  port: 8002                                                 #本服务端口
  readTimeout: 5000                                          #fasthttp client完整响应读取(包括正文)的最大持续时间
  writeTimeout: 5000                                         #fasthttp client完整请求写入(包括正文)的最大持续时间
  noDefaultUserAgentHeader: true
  maxConnPerHost: 10000
  MaxIdleConnDuration: 5000                                  #空闲的保持连接将在此持续时间后关闭
  NoDefaultUserAgentHeader: 30000

redis:
  address: ""                                                #redis地址
  password: "apipost"
  db: 1

kafka:                                                           
  address: ""                                                #kafka地址
  topIc: "report"


mongo:
  dsn: ""                                                    #mongo地址
  database: "runnergo_open"                                  #mongo库
  stressDebugTable: "stress_debug"                           # 性能测试中接口debug日志集合
  sceneDebugTable: "scene_debug"                             # 场景调试，接口debug日志集合
  apiDebugTable: "api_debug"                                  # api调试，接口debug日志集合
  debugTable: "debug_status"                                  # 性能测试中，是否开启debug日志模式集合
  autoTable: "auto_report"                                   # 自动化测试，接口debug日志集合



machine:
  maxGoroutines: 20005                                      
  serverType: 1
  netName: ""
  diskName: ""


log:
  path: "/data/logs/RunnerGo/RunnerGo-engine-info.log"         #本服务log存放地址


management:
  notifyStopStress: "https://****/management/api/v1/plan/notify_stop_stress"                          #management服务停止性能任务接口
  notifyRunFinish: "https://***/management/api/v1/auto_plan/notify_run_finish"                           #management服务任务完成接口
```


# 启动
```text
 配置完成后，在根目录./main启动engine服务
```

## 开源部署
1. 配置环境变量
## 配置说明
| key                               | 是否必填 | 默认值                                                                       |                                   说明 |
|:----------------------------------|------|---------------------------------------------------------------------------|-------------------------------------:|
| mongo数据库                          ||||
| RG_MONGO_DSN                      | 否    | 默认：mongodb://runnergo:123456@127.0.0.0:27017/runnergo                     |                          mongo数据库dsn |
| RG_MONGO_DATABASE                 | 否    | 默认：runnergo                                                               |                          mongo使用的哪个库 |
| RG_MONGO_STRESS_DEBUG_TABLE       | 否    | 默认：stressDebugTable                                                       |                     性能测试debug日志存储的集合 |
| RG_MONGO_DEBUG_TABLE              | 否    | 默认：debugTable                                                             |                        性能测试debug模式状态 |
| RG_MONGO_SCENE_DEBUG_TABLE        | 否    | 默认：sceneDebugTable                                                        |                          场景调试日志存储的集合 |
| RG_MONGO_API_DEBUG_TABLE          | 否    | 默认：apiDebugTable                                                          |                          接口调试日志存储的集合 |
| RG_MONGO_AUTO_TABLE               | 否    | 默认：auto_table                                                             |                           自动化日志存储的集合 |
| Redis                             ||||
| RG_REDIS_ADDRESS                  | 否    | 默认：127.0.0.0:6379                                                         |                           redis服务端地址 |
| RG_REDIS_PASSWORD                 | 是    |                                                                           |                           redis服务端密码 |
| RG_DB                             | 否    | 默认：0                                                                      |                             redis数据库 |
| kafka配置                           |      |                                                                           |                                      |
| RG_KAFKA_TOPIC                    | 否    | 默认：runnergo                                                               |                          kafka的topic |
| RG_KAFKA_ADDRESS                  | 否    |                                                                           |                              kafka地址 |
| RG_ENGINE_HTTP_NAME               | 否    |                                                                           |                                      |
| RG_ENGINE_HTTP_ADDRESS            | 否    | 0.0.0.0:30000                                                             |                           engine服务地址 |
| 本服务和使用的httpclient                 |      |                                                                           |                                      |
| HTTP_VERSION                      | 否    |                                                                           |                                      |
| HTTP_READ_TIMEOUT                 | 否    | 默认5000毫秒                                                                  |                  完整响应读取(包括正文)的最大持续时间 |
| HTTP_WRITE_TIMEOUT                | 否    | 默认5000毫秒                                                                  |                  完整请求写入(包括正文)的最大持续时间 |                      |      |                                              |                      |
| HTTP_MAX_CONN_PER_HOST            | 否    | 默认10000                                                                   |                       每台主机可以建立的最大连接数 |
| HTTP_MAX_IDLE_CONN_DURATION       | 否    | 默认5000毫秒                                                                  |                    空闲的保持连接将在此持续时间后关闭 |
| HTTP_MAX_CONN_WAIT_TIMEOUT        | 否    |                                                                           |                                      |
| HTTP_NO_DEFAULT_USER_AGENT_HEADER | 否    |                                                                           |                                      |
| RG_COLLECTOR_HTTP_HOST            | 否    | 默认：0.0.0.0:30000                                                          |                                      |
| 日志文件地址                            |      |                                                                           |                                      |
| RG_ENGINE_LOG_PATH                | 否    | /data/logs/RunnerGo/RunnerGo-engine-info.log                              |                               日志文件地址 |
| management服务                      |      |                                                                           |                                      |
| RG_MANAGEMENT_NOTIFY_STOP_STRESS  | 否    | 默认： https://127.0.0.0:30000/management/api/v1/plan/notify_stop_stress     |                 management服务地址停止任务接口 |
| RG_MANAGEMENT_NOTIFY_RUN_FINISH   | 否    | 默认： https://127.0.0.0:30000/management/api/v1/auto_plan/notify_run_finish |                 management服务地址完成任务接口 |
| 本机设置                              |      |                                                                           |                                      |
| RG_MACHINE_MAX_GOROUTINES         | 否    | 默认：20000                                                                  |                             最大支持协程数字 |
| RG_MACHINE_SERVER_TYPE            | 否    | 默认：备用机                                                                    |                              是否为备用机器 |
| RG_MACHINE_NET_NAME               | 否    |                                                                           |                            本机使用的网络名称 |
| RG_MACHINE_DISK_NAME              | 否    |                                                                           |                            本机使用的磁盘名称 |
| 心跳配置                              |      |                                                                           |                                      |
| RG_HEARTBEAT_PORT                 | 否    | 默认：30000                                                                  |                                本服务端口 |
| RG_HEARTBEAT_REGION               | 否    | 默认：北京                                                                     |                               本机所在城市 |
| RG_HEARTBEAT_DURATION             | 否    | 默认3秒                                                                      |                         多长时间发送一次心跳数据 |
| RG_HEARTBEAT_RESOURCES            | 否    | 默认3秒                                                                      |                     多长时间发送一次本机资源使用数据 |


