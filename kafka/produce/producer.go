// Package produce @Author  wangjian    2023/7/21 12:25 AM
package produce

import (
	"github.com/JianWangEx/commonService/kafka/config"
	"github.com/Shopify/sarama"
)

var (
	kafkaProducerMap = make(map[string]sarama.SyncProducer)
)

type KafkaMessage struct {
	Topic string      `json:"topic"`
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type ProducerError struct {
	Partition int32 `json:"partition"`
	Offset    int64 `json:"offset"`
	error
}

func NewProducerError(partition int32, offset int64, err error) *ProducerError {
	return &ProducerError{
		Partition: partition,
		Offset:    offset,
		error:     err,
	}
}

//func SendMessage(ctx context.Context, msg *KafkaMessage) error {
//
//}

func GetKafkaProducerMap() (map[string]sarama.SyncProducer, error) {
	if len(kafkaProducerMap) == 0 {
		err := initKafkaProducerMap()
		if err != nil {
			return nil, err
		}
	}
	return kafkaProducerMap, nil
}

func initKafkaProducerMap() error {
	m := config.GetKafkaClusterConfigMap()
	for k, v := range m {
		producer, err := sarama.NewSyncProducer(v.Brokers, config.GetDefaultKafkaConfig())
		if err != nil {
			kafkaProducerMap = map[string]sarama.SyncProducer{}
			return err
		}
		kafkaProducerMap[k] = producer
	}
	return nil
}
