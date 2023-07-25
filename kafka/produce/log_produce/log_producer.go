// Package log_produce @Author  wangjian    2023/7/22 4:25 PM
package log_produce

import (
	"context"
	"github.com/JianWangEx/commonService/kafka/produce"
	logger "github.com/JianWangEx/commonService/log"
	"github.com/JianWangEx/commonService/util"
	"github.com/Shopify/sarama"
)

var logProducer *LogProducer

type LogProducer struct {
	syncProducer  sarama.SyncProducer
	asyncProducer sarama.AsyncProducer
}

func (p *LogProducer) SendSaramaMessage(ctx context.Context, message *produce.KafkaMessage) error {
	msg := &sarama.ProducerMessage{
		Topic: message.Topic,
		Key:   sarama.StringEncoder(util.SafeToJson(message.Key)),
		Value: sarama.StringEncoder(util.SafeToJson(message.Value)),
	}
	partition, offset, err := p.syncProducer.SendMessage(msg)
	if err != nil {
		logger.CtxSugar(ctx).Errorf("[log_producer]SendSaramaMessage|sending message failed, message: %v, error: %v", *message, err)
	}
	logger.CtxSugar(ctx).Infof("[log_producer]SendSaramaMessage|sending message success, message: %v, partition: %d, offset: %d", *message, partition, offset)
	return nil
}

func GetLogProducer() *LogProducer {
	return logProducer
}

func InitLogProducer() error {
	m, err := produce.GetKafkaProducerMap()
	if err != nil {
		return err
	}
	logProducer = &LogProducer{
		syncProducer: m["log"],
	}
	return nil
}
