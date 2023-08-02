// Package produce @Author  wangjian    2023/7/21 9:44 AM
package produce

import (
	"context"
	"github.com/Shopify/sarama"
)

type Client interface {
	SendSaramaMessage(ctx context.Context, message *sarama.ProducerMessage) error
}

func GetClient() Client {
	return sendManager
}
