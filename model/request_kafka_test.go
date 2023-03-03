package model

import (
	"fmt"
	"github.com/Shopify/sarama"
	"testing"
)

func TestSendKafkaMsg(t *testing.T) {
	address := "101.201.109.168:9092"
	//_, err := NewKafkaProducer([]string{address})
	//if err != nil {
	//	fmt.Println("kafka连接失败", err)
	//	return
	//}
	//
	//fmt.Println("kafk:    ", err)

	//resultDataMsgCh := make(chan *ResultDataMsg, 100)
	//
	//resultDataMsg := new(ResultDataMsg)
	//
	//for i := 0; i < 100; i++ {
	//	resultDataMsg.TargetId = fmt.Sprintf("%d", i)
	//	resultDataMsgCh <- resultDataMsg
	//	fmt.Println(resultDataMsg)
	//}
	//wg := &sync.WaitGroup{}
	//wg.Add(1)
	//ip := ""
	//go SendKafkaMsg(kafkaProducer, resultDataMsgCh, "StressTestData", 1, ip)
	//wg.Wait()
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Return.Errors = true
	_, err := sarama.NewConsumer([]string{address}, saramaConfig)
	fmt.Println("err:   ", err)

}
