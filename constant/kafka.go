// Package constant @Author  wangjian    2023/8/1 6:40 PM
package constant

const (
	KafkaHeaderKeyGroup      = "group"
	KafkaHeaderKeyTopic      = "topic"
	KafkaHeaderKeyTraceId    = "traceId"
	KafkaHeaderKeyRetryTimes = "retryTimes"

	DefaultKafkaProducerClusterName = "default_producer"
	DefaultKafkaConsumerClusterName = "default_consumer"
)

const (
	DefaultKafkaConsumingGoroutines = 8
	MaxKafkaConsumingGoroutines     = 100
)

const (
	ConsumerGroupYoga = "yoga"
)

var (
	DefaultGroup = make(map[string]string)
	YogaGroup    = make(map[string]string)
)

const (
	halfMinute    = uint32(30)
	oneMinute     = uint32(60)
	twoMinute     = uint32(120)
	threeMinute   = uint32(180)
	fiveMinute    = uint32(300)
	tenMinute     = uint32(600)
	fifteenMinute = uint32(900)
	halfHour      = uint32(1800)
	oneHour       = uint32(3600)
	twoHour       = uint32(7200)
	fourHour      = uint32(14400)
	sixHour       = uint32(21600)
	halfDay       = uint32(43200)
	oneDay        = uint32(86400)
)

var DelayTimes = []uint32{halfMinute, oneMinute, twoMinute, threeMinute, fiveMinute, tenMinute, fifteenMinute,
	halfHour, oneHour, twoHour, fourHour, sixHour, halfDay, oneDay}
