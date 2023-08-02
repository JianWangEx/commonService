// Package log_produce @Author  wangjian    2023/7/22 6:39 PM
package log_produce

import (
	"context"
	"github.com/JianWangEx/commonService/kafka/config"
	"github.com/JianWangEx/commonService/kafka/produce"
	"testing"
)

func TestProduce(t *testing.T) {
	path := "./../../config/config_test.toml"
	config.InitKafkaConfig()
	err := config.InitKafkaClusterConfig(path)
	if err != nil {
		panic(err)
	}

	err = InitLogProducer()
	if err != nil {
		panic(err)
	}

	message := &produce.KafkaMessage{
		Topic:       "kafka_test",
		MessageBody: "kafka_test_message",
	}

	err = GetLogProducer().SendSaramaMessage(context.TODO(), message)
	if err != nil {
		panic(err)
	}
}
