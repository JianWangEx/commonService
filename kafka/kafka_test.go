// Package kafka @Author  wangjian    2023/8/2 11:23 PM
package kafka

import (
	"context"
	"github.com/JianWangEx/commonService/constant"
	"github.com/JianWangEx/commonService/kafka/config"
	"github.com/JianWangEx/commonService/kafka/consume"
	"github.com/JianWangEx/commonService/kafka/produce"
	"testing"
	"time"
)

func TestKafka(t *testing.T) {
	topic := "test_log"

	path := "./config/config_test.toml"
	config.InitKafkaConfig()
	err := config.InitKafkaClusterConfig(path)
	if err != nil {
		panic(err)
	}

	err = produce.ClientInit()
	if err != nil {
		panic(err)
	}

	ctx := context.TODO()
	consume.RegisterKafkaConsumer(ctx)

	go func() {
		for i := 0; i < 10; i++ {
			err = produce.SendKafkaMessage(ctx, &produce.KafkaMessage{
				Topic:                 topic,
				Group:                 constant.KafkaGroupDefault,
				DelaySendTimeInternal: 10,
				MessageBody:           "test kafka message",
			})
			if err != nil {
				panic(err)
			}
			time.Sleep(1 * time.Second)
		}
		time.Sleep(time.Minute)
	}()

	time.Sleep(time.Minute * 5)
}
