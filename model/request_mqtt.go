package model

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	uuid "github.com/satori/go.uuid"
)

type MQTT struct {
	TargetId       string          `json:"target_id" bson:"target_id"`
	Uuid           uuid.UUID       `json:"uuid" bson:"uuid"`
	Name           string          `json:"name" bson:"name"`
	TeamId         string          `json:"team_id" bson:"team_id"`
	TargetType     string          `json:"target_type" bson:"target_type"` // api/webSocket/tcp/grpc
	Timeout        int64           `json:"timeout" bson:"timeout"`         // 请求超时时间
	Debug          string          `json:"debug" bson:"debug"`             // 是否开启Debug模式
	MQTTConfig     MQTTConfig      `json:"mqtt_config"`
	Configuration  *Configuration  `json:"configuration" bson:"configuration"`
	MqttVariable   *GlobalVariable `json:"mqtt_variable"`   // 全局变量
	GlobalVariable *GlobalVariable `json:"global_variable"` // 全局变量
}

type MQTTConfig struct {
	PortType     string       `json:"port_type"` // 端口类型：tcp\ws\wss\ssl， 默认tcp
	Broker       string       `json:"broker"`    // 必填项
	Port         int          `json:"port"`      // 端口
	Action       string       `json:"action"`
	Topic        string       `json:"topic"` // 必填项
	CommonConfig CommonConfig `json:"common_config"`
	HigherConfig HigherConfig `json:"higher_config"`
	PingTimeOut  int64        `json:"ping_time_out"` // 非必填

	WriteTimeOut     int64   `json:"write_time_out"`     // 非必填
	KeepLiveTime     int64   `json:"keep_live_time"`     // 默认10 非必填
	MQTTVersion      uint    `json:"mqtt_version"`       // 版本 默认5.0
	AutoReconnect    bool    `json:"auto_reconnect"`     // 是否自动重连
	Will             Will    `json:"will"`               // 遗愿
	Retained         bool    `json:"retained"`           // 是否保留消息
	Qos              byte    `json:"qos"`                // 服务质量： 0：至多一次 ； 1：至少一次； 2：确保只有一次
	IsEncrypt        bool    `json:"is_encrypt"`         // 是否认证
	Tls              Tls     `json:"tls"`                // tls 认证
	CaFile           *CaFile `json:"ca_file"`            // 认证文件
	OnConnectionLost bool    `json:"on_connection_lost"` // 断开连接后是否需要重连
}

type CommonConfig struct {
	Username string `json:"username"`  // 非必填
	Password string `json:"password"`  // 非必填
	ClientId string `json:"client_id"` // 必填
}

type HigherConfig struct {
	ConnectTimeOut int64 `json:"connect_time_out"` // 非必填
}
type Will struct {
	WillEnabled bool   //遗愿
	WillTopic   string //遗愿主题
	WillPayload string //遗愿消息
	WillQos     byte   //遗愿服务质量
}
type Tls struct {
	IsTls    bool `json:"is_tls"`    // 是否使用Tls链接
	IsClient bool `json:"is_client"` // 客户端是否需要证书
}

type CaFile struct {
	CaPem        string `json:"ca_pem"`
	ClientCrtPem string `json:"client_crt_pem"`
	ClientKeyPem string `json:"client_key_pem"`
}

type MQTTClient struct {
	Client   mqtt.Client                    `json:"client"`    // MQTT客户端
	Topics   map[string]mqtt.MessageHandler `json:"topics"`    // topic
	Retained bool                           `json:"retained"`  //
	QOS      byte                           `json:"qos"`       // 质量
	IsRepeat bool                           `json:"is_repeat"` // 是否重复发送
}

// MessagePubHandler 创建全局mqtt publish消息处理 handler
var MessagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Sprintf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

/**
 *  @Description: 发布消息
 *  @param: top 主题， msgList 要发布的消息
 */

func (c *MQTTClient) Publish(topic string, msgList []string) {
	if c.Client == nil || !c.Client.IsConnected() {
		return
	}
	for _, v := range msgList {
		if token := c.Client.Publish(topic, c.QOS, c.Retained, v); token.Wait() && token.Error() != nil {
			continue
		}
	}
}

// Subscribe 订阅
func (c *MQTTClient) Subscribe(topics []string) {
	if c.Client == nil || !c.Client.IsConnected() {
		return
	}
	for _, topic := range topics {
		if token := c.Client.Subscribe(topic, c.QOS, MessagePubHandler); token.Wait() && token.Error() != nil {
			continue
		}
		c.Topics[topic] = MessagePubHandler
	}
}

// Close 关闭mqtt客户端
func (c *MQTTClient) Close() {
	if c.Client == nil {
		return
	}
	c.Client.Disconnect(250)
}
