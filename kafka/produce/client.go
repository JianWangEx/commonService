// Package produce @Author  wangjian    2023/7/21 9:44 AM
package produce

import (
	"context"
)

type Client interface {
	SendSaramaMessage(ctx context.Context, message *KafkaMessage) error
}
