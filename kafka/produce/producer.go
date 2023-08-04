// Package produce @Author  wangjian    2023/7/21 12:25 AM
package produce

import (
	"context"
	"github.com/JianWangEx/commonService/constant"
	"github.com/JianWangEx/commonService/kafka/config"
	"github.com/JianWangEx/commonService/kafka/delay"
	logger "github.com/JianWangEx/commonService/log"
	"github.com/JianWangEx/commonService/util"
	"github.com/Shopify/sarama"
	"strconv"
	"time"
)

var (
	kafkaProducerMap = make(map[string]sarama.SyncProducer)

	sendManager = &SendManager{}
)

type KafkaMessage struct {
	// the topic to send message
	Topic string
	// the name of consumer group that the message needs to be consumed
	Group string
	// the time for delay send, the uint is seconds. if zero, send immediately
	DelaySendTimeInternal uint32
	// the message body
	MessageBody interface{}
}

func ClientInit() error {
	var initErr error
	// init default producer
	defaultSarama := config.Kafka().Sarama
	commonCluster := config.KafkaCluster{
		Name:   constant.DefaultKafkaProducerClusterName,
		Sarama: defaultSarama,
	}
	err := singleClientInit(commonCluster)
	if err != nil {
		initErr = err
	}

	for _, singleCluster := range config.Kafka().ProducerCluster {
		err = singleClientInit(singleCluster)
		if err != nil {
			initErr = err
		}
	}

	return initErr
}

func singleClientInit(cluster config.KafkaCluster) error {
	saramaConfig := config.GetDefaultKafkaConfig()
	client, err := sarama.NewClient(cluster.Brokers, saramaConfig)
	if err != nil {
		return err
	}
	kafkaSyncProducer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return err
	}
	kafkaProducerMap[cluster.Name] = kafkaSyncProducer
	return nil

}

// SendKafkaMessage send kafka message
func SendKafkaMessage(ctx context.Context, msg *KafkaMessage) error {
	onceLog := logger.CtxSugar(ctx)
	saramaMsg, err := generateProducerMessage(ctx, msg)
	if err != nil {
		onceLog.Errorf("kafka generate produce message err: %+v, msg: %+v", err, msg)
		return err
	}
	return GetClient().SendSaramaMessage(ctx, saramaMsg)
}

// SendKafkaConsumeMessage send consume message
func SendKafkaConsumeMessage(ctx context.Context, consumeMsg *sarama.ConsumerMessage, delayTime uint32) error {
	onceLog := logger.CtxSugar(ctx)
	saramaMsg, err := generateSaramaMsgConsume(ctx, consumeMsg, delayTime)
	if err != nil {
		onceLog.Errorf("kafka generate consume produce message err: %+v, msg: %+v", err, consumeMsg)
		return err
	}
	return GetClient().SendSaramaMessage(ctx, saramaMsg)
}

type SendManager struct{}

func (m *SendManager) SendSaramaMessage(ctx context.Context, message *sarama.ProducerMessage) error {
	onceLog := logger.CtxSugar(ctx)
	start := time.Now()
	var syncProducer sarama.SyncProducer
	// get sync producer by message topic
	topicToClusterMap := config.GetProducerTopicToClusterMap()
	if clusterName, ok := topicToClusterMap[message.Topic]; ok {
		if producer, ok := kafkaProducerMap[clusterName]; ok {
			syncProducer = producer
		}
	}
	if syncProducer == nil {
		if defaultProducer, ok := kafkaProducerMap[constant.DefaultKafkaProducerClusterName]; ok {
			syncProducer = defaultProducer
		}
	}
	if syncProducer == nil {
		return constant.KafkaErrorClientNilErr
	}

	partition, offset, err := syncProducer.SendMessage(message)
	if err != nil {
		onceLog.Errorf("kafka send msg err: %+v, msg: %+v", err, message)
		// TODO: add monitor report
		return err
	}
	cost := time.Since(start)
	onceLog.Infof("kafka send message success, partition: %+v, offset: %+v, topic: %+v, targetTopic: %+v, msg: %+v, cost: +%v", partition, offset, message.Topic, getStrFromProducerMsgHeader(message,
		constant.KafkaHeaderKeyTopic), message.Value, cost)
	// TODO: add monitoring
	return nil
}

func generateProducerMessage(ctx context.Context, msg *KafkaMessage) (*sarama.ProducerMessage, error) {
	if msg.Group == "" {
		return nil, constant.KafkaErrorGroupEmpty
	}
	saramaMsg := new(sarama.ProducerMessage)
	saramaMsg.Topic = msg.Topic
	saramaMsg.Value = sarama.StringEncoder(util.SafeToJson(msg.MessageBody))
	addHeaderInfo(saramaMsg, constant.KafkaHeaderKeyGroup, msg.Group)
	addHeaderInfo(saramaMsg, constant.KafkaHeaderKeyTopic, msg.Topic)
	addHeaderInfo(saramaMsg, constant.KafkaHeaderKeyTraceId, logger.GetTraceIDFromCtx(ctx))
	addHeaderInfo(saramaMsg, constant.KafkaHeaderKeyRetryTimes, strconv.Itoa(0))
	return saramaMsg, nil
}

func addHeaderInfo(msg *sarama.ProducerMessage, key, value string) {
	recordHeader := new(sarama.RecordHeader)
	recordHeader.Key = []byte(key)
	recordHeader.Value = []byte(value)
	msg.Headers = append(msg.Headers, *recordHeader)
}

func getStrFromProducerMsgHeader(msg *sarama.ProducerMessage, key string) string {
	for _, header := range msg.Headers {
		if string(header.Key) == key {
			return string(header.Value)
		}
	}
	return ""
}

func generateSaramaMsgConsume(ctx context.Context, consumeMsg *sarama.ConsumerMessage, delayTime uint32) (*sarama.ProducerMessage, error) {
	kafkaMsg := new(sarama.ProducerMessage)
	kafkaMsg.Topic = generateTopic(ctx, consumeMsg.Topic, delayTime)
	kafkaMsg.Value = sarama.StringEncoder(consumeMsg.Value)
	for _, v := range consumeMsg.Headers {
		kafkaMsg.Headers = append(kafkaMsg.Headers, *v)
	}
	return kafkaMsg, nil
}

func generateTopic(ctx context.Context, topic string, delayTime uint32) string {
	return delay.GetDelayTopic(delayTime, topic)
}
