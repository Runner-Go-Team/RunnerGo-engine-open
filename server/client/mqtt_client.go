package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"io/ioutil"
	"time"
)

// NewMqttClient /**
func NewMqttClient(config model.MQTTConfig) *model.MQTTClient {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("%s://%s:%d", config.PortType, config.Broker, config.Port))
	opts.SetClientID(config.ClientId)
	if config.Username != model.NILSTRING {
		opts.SetUsername(config.Username)
	}

	if config.Password != model.NILSTRING {
		opts.SetPassword(config.Password)
	}
	if config.PingTimeOut != model.NILINT {
		opts.SetPingTimeout(time.Duration(config.PingTimeOut) * time.Second)
	}
	if config.ConnectTimeOut != model.NILINT {
		opts.SetConnectTimeout(time.Duration(config.ConnectTimeOut) * time.Second)
	}
	if config.WriteTimeOut != model.NILINT {
		opts.SetWriteTimeout(time.Duration(config.WriteTimeOut) * time.Second)
	}
	if config.KeepLiveTime != model.NILINT {
		opts.SetKeepAlive(time.Duration(config.KeepLiveTime) * time.Second)
	}
	if config.MQTTVersion != model.NILINT {
		opts.SetProtocolVersion(config.MQTTVersion)
	}

	if config.Tls.IsTls && config.CaFile != nil {
		opts.SetTLSConfig(newTlsConfig(config.CaFile, config.Tls.IsClient))
	}

	if config.Will.WillEnabled {
		opts.SetWill(config.Will.WillTopic, config.Will.WillPayload, config.Will.WillQos, config.Retained)
	}
	opts.OnConnect = func(client mqtt.Client) {
		//log.Logger.Debug("MQTT连接成功")
		fmt.Println("mqtt连接成功")
	}
	if config.OnConnectionLost {
		opts.OnConnectionLost = func(client mqtt.Client, err error) {
			log.Logger.Debug("MQTT连接断开")
			if config.AutoReconnect {
				NewMqttClient(config)
			}
		}
	}
	opts.SetDefaultPublishHandler(model.MessagePubHandler)
	client := mqtt.NewClient(opts)
	c := new(model.MQTTClient)
	c.Client = client
	c.QOS = config.Qos
	c.Retained = config.Retained
	c.Topics = make(map[string]mqtt.MessageHandler)
	if tc := c.Client.Connect(); tc.Wait() && tc.Error() != nil {
		log.Logger.Debug("mqtt 建立连接失败!")
		return nil
	}
	return c
}

func newTlsConfig(caFile *model.CaFile, isClient bool) (tlsConfig *tls.Config) {
	if caFile == nil {
		return
	}
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caFile.CaPem)
	if err != nil {
		log.Logger.Error("pem文件打开失败")
		return
	}
	certPool.AppendCertsFromPEM(ca)
	if isClient {
		// Import client certificate/key pair
		clientKeyPair, err := tls.LoadX509KeyPair(caFile.ClientCrtPem, caFile.ClientKeyPem)
		if err != nil {
			log.Logger.Error("certFile或keyFile文件使用失败")
			return
		}
		tlsConfig = &tls.Config{
			RootCAs:            certPool,
			ClientAuth:         tls.NoClientCert,
			ClientCAs:          nil,
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{clientKeyPair},
		}
	} else {
		tlsConfig = &tls.Config{
			RootCAs: certPool,
		}
	}

	return tlsConfig
}
