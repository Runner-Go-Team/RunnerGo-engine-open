package model

import (
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Shopify/sarama"
)

/*
 将需要的测试数据写入到kafka中
*/

// SendKafkaMsg 发送消息到kafka
func SendKafkaMsg(kafkaProducer sarama.SyncProducer, resultDataMsgCh chan *ResultDataMsg, topic string, partition int32, ip string) {
	defer kafkaProducer.Close()
	teamId, planId, reportId, num, index := "", "", "", int64(0), 0
	for {
		if resultDataMsg, ok := <-resultDataMsgCh; ok {
			msg, err := json.Marshal(&resultDataMsg)
			if err != nil {
				log.Logger.Error(fmt.Sprintf("机器ip:%s, json转换失败: %s", ip, err))
				break
			}
			if num == 0 {
				num = resultDataMsg.MachineNum
			}
			index++
			reportId = resultDataMsg.ReportId
			planId = resultDataMsg.PlanId
			teamId = resultDataMsg.TeamId
			DataMsg := &sarama.ProducerMessage{}
			DataMsg.Topic = topic
			DataMsg.Partition = partition
			DataMsg.Value = sarama.StringEncoder(msg)
			_, _, err = kafkaProducer.SendMessage(DataMsg)
			if err != nil {
				log.Logger.Error(fmt.Sprintf("机器ip:%s, 向kafka发送消息失败: %s", ip, err.Error()))
				break
			}
		} else {
			//// 发送结束消息
			//SendStartStopMsg(kafkaProducer, topic, partition, reportId, num)
			log.Logger.Info(fmt.Sprintf("机器ip: %s, 团队: %s，计划: %s，报告: %s, 测试数据向kafka写入完成！本次任务有： %d 条数据", ip, teamId, planId, reportId, index))
			return

		}

	}
}

// NewKafkaProducer 构建生产者
func NewKafkaProducer(addrs []string) (kafkaProducer sarama.SyncProducer, err error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll           // 发送完数据需要leader和follow都确认
	config.Producer.Partitioner = sarama.NewManualPartitioner  // 设置选择分区的策略为Hash,当设置key时，所有的key的消息都在一个分区Partitioner里
	config.Producer.Return.Successes = true                    // 成功交付的消息将在success channel返回
	kafkaProducer, err = sarama.NewSyncProducer(addrs, config) // 生产者客户端
	if err != nil {
		return
	}
	return
}
