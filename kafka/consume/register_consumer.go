// Package consume @Author  wangjian    2023/8/1 6:31 PM
package consume

import (
	"context"
	"errors"
	"github.com/JianWangEx/commonService/constant"
	"github.com/JianWangEx/commonService/kafka/config"
	logger "github.com/JianWangEx/commonService/log"
	"github.com/Shopify/sarama"
	"strings"
	"time"
)

var (
	consumeFuncMap = make(map[string]KafkaConsumeFunc)
)

func RegisterKafkaConsumer(ctx context.Context) {
	initConsumerFunc()
	for _, consumer := range config.Kafka().Consumers {
		switch strings.ToLower(strings.TrimSpace(consumer.GroupLevel)) {
		case constant.ConsumerGroupYoga:
			registerConsumer(ctx, consumer, constant.YogaGroup)
		default:
			registerConsumer(ctx, consumer, constant.DefaultGroup)
		}
	}
}

func registerConsumer(ctx context.Context, consumer config.Consumer, groupMap map[string]string) {
	logger.CtxSugar(ctx).Debugf("[registerConsumer]consumer=%v, groupMap=%v", consumer, groupMap)
	consumeFunc, ok := getRealConsumeFunc(ctx, consumer)
	if !ok {
		logger.CtxSugar(ctx).Warnf("[register_consumer]getRealConsumeFunc kafka consume func not found, topic=%s", consumer.Topic)
		return
	}
	consumerConfig := new(ConsumerConfig)
	consumerConfig.KafkaConsumeFunc = consumeFunc
	consumerConfig.Topic = consumer.Topic
	consumerConfig.DelayTime = consumer.DelayTime
	consumerConfig.RetryTimes = consumer.RetryTimes
	consumerConfig.ConcurrentNums = consumer.ConcurrentNums
	for _, group := range groupMap {
		consumerConfig.GroupId = group
		doRegisterKafkaConsumer(ctx, *consumerConfig)
	}
}

func initConsumerFunc() {}

// getRealConsumeFunc get real consumer function
func getRealConsumeFunc(ctx context.Context, consumer config.Consumer) (KafkaConsumeFunc, bool) {
	f, ok := consumeFuncMap[consumer.Topic]
	if !ok {
		logger.CtxSugar(ctx).Warnf("[register_consumer]getRealConsumeFunc kafka consume func not found, topic=%s", consumer.Topic)
		return nil, false
	}
	return f, true
}

func doRegisterKafkaConsumer(ctx context.Context, consumerConfig ConsumerConfig) {
	if consumerConfig.RetryTimes > 0 && len(consumerConfig.DelayTime) == 0 {
		registerErr := errors.New("register param err, retryTimes is not zero while delayTime is empty")
		panic(registerErr)
	}
	// get consumer cluster name to cluster mapping
	consumerClusterMap := config.GetKafkaConsumerClusterMap()
	consumerCluster := consumerClusterMap[constant.DefaultKafkaConsumerClusterName]
	// get consumer topic to cluster name mapping
	topicClusterMap := config.GetConsumerTopicToClusterMap()
	if clusterName, ok := topicClusterMap[consumerConfig.Topic]; ok {
		if clusterConfig, ok := consumerClusterMap[clusterName]; ok {
			consumerCluster = clusterConfig
		}
	}

	kafkaDefaultConfig := config.GetDefaultKafkaConfig()
	kafkaClient, err := sarama.NewClient(consumerCluster.Brokers, kafkaDefaultConfig)
	if err != nil {
		logger.CtxSugar(ctx).Warnf("[register_consumer]doRegisterKafkaConsumer sarama new client failed: %v", err)
		panic(err)
	}

	consumerGroup, err := sarama.NewConsumerGroupFromClient(consumerConfig.GroupId, kafkaClient)
	if err != nil {
		logger.CtxSugar(ctx).Warnf("[register_consumer]doRegisterKafkaConsumer sarama new consumer group from client failed: %v", err)
		panic(err)
	}

	brokers := consumerCluster.Brokers
	go func(ConsumerConfig, sarama.ConsumerGroup, []string) {
		consumerHandler := NewKafkaConsumer(consumerConfig)
		logger.CtxSugar(ctx).Infof("register kafka consumer, config: %+v", consumerHandler)
		for {
			err = consumerGroup.Consume(ctx, []string{consumerConfig.Topic}, consumerHandler)
			if err != nil {
				logger.CtxSugar(ctx).Errorf("kafka consumer register failed: %+v, topic: %+v, brokers: %+v", err, consumerConfig.Topic, brokers)
			}
			if errors.Is(err, sarama.ErrUnknownTopicOrPartition) {
				time.Sleep(time.Minute * 3)
			} else {
				time.Sleep(time.Second * 3)
			}
		}
	}(consumerConfig, consumerGroup, brokers)
}
