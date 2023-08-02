// Package consume @Author  wangjian    2023/7/22 6:45 PM
package consume

import (
	"context"
	"fmt"
	"github.com/JianWangEx/commonService/constant"
	"github.com/JianWangEx/commonService/kafka/produce"
	logger "github.com/JianWangEx/commonService/log"
	"github.com/Shopify/sarama"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var emptyStruct = struct{}{}

// KafkaConsumeFunc func to consume Kafka messages
type KafkaConsumeFunc func(context.Context, string, []*sarama.RecordHeader) error

type ConsumerConfig struct {
	// consumer group, the same message in group will be consumed only once for one group
	GroupId string `json:"groupId"`
	// the topic of message to consume
	Topic string `json:"topic"`
	// when need retry with different time internal, use this when failed to reconsume message
	DelayTime []uint32
	// the count for retry times
	RetryTimes uint32 `json:"retryTimes"`
	// the number of concurrent consumption
	ConcurrentNums uint32
	// consume func
	KafkaConsumeFunc
}

type DataSyncConsumer struct {
	ConsumerConfig
	// partition info, partition number to partition info mapping
	consumingInfo map[int32]*partitionConsumingInfo

	// Used to control the number of concurrency with the number of partitions as the dimension
	chanMap map[int32]chan interface{}
	sync.Mutex
}

// partitionConsumingInfo the consuming message info
type partitionConsumingInfo struct {
	// offset to message mapping
	consumingMap map[int64]*sarama.ConsumerMessage
	// the max offset on consuming
	maxOffset int64
	// the min offset on consuming
	minOffset int64
}

func NewKafkaConsumer(consumerConfig ConsumerConfig) *DataSyncConsumer {
	consumer := new(DataSyncConsumer)
	if consumerConfig.ConcurrentNums <= 0 {
		consumerConfig.ConcurrentNums = constant.DefaultKafkaConsumingGoroutines
	}
	if consumerConfig.ConcurrentNums > constant.MaxKafkaConsumingGoroutines {
		consumerConfig.ConcurrentNums = constant.MaxKafkaConsumingGoroutines
	}
	consumer.ConsumerConfig = consumerConfig

	consumer.consumingInfo = make(map[int32]*partitionConsumingInfo)
	consumer.chanMap = make(map[int32]chan interface{})
	return consumer
}

func (c *DataSyncConsumer) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *DataSyncConsumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *DataSyncConsumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		if cap(c.chanMap[msg.Partition]) == 0 {
			c.chanMap[msg.Partition] = make(chan interface{}, c.ConcurrentNums)
		}

		// check the group for msg
		group := getStrFromMsgHeader(msg, constant.KafkaHeaderKeyGroup)
		if group != "" && !strings.HasPrefix(c.GroupId, group) {
			// if no msg is being consumed, the offset is submitted directly
			if len(c.chanMap[msg.Partition]) == 0 {
				sess.MarkMessage(msg, "")
			}
			continue
		}
		c.beforeConsume(msg)
		ctx := generateMsgCtx(msg)
		onceLog := logger.CtxSugar(ctx)

		msgLatency := time.Since(msg.Timestamp)
		// TODO: add monitor report
		onceLog.Infof("start consume offset: %+v, partition: %+v, topic: %+v, latency: +%v, value: +%v", msg.Offset, msg.Partition, msg.Topic, msgLatency, msg.Value)

		go c.process(ctx, sess, msg)

	}
	return nil
}

// beforeConsume: add message to c.consumingInfo[message.Partition].consumingMap[message.Offset] or init c.consumingInfo[message.Partition]
func (c *DataSyncConsumer) beforeConsume(message *sarama.ConsumerMessage) {
	c.chanMap[message.Partition] <- emptyStruct
	// need to add lock, avoid concurrent errors in simultaneous processing of multiple covariances
	c.Lock()
	defer c.Unlock()
	partitionInfo, ok := c.consumingInfo[message.Partition]
	if ok {
		partitionInfo.consumingMap[message.Offset] = message
		// ensure atomic update, avoid data race issues that can occur when multiple goroutines are accessed concurrently
		atomic.StoreInt64(&partitionInfo.maxOffset, message.Offset)
		if len(c.chanMap[message.Partition]) == 1 || partitionInfo.minOffset == 0 {
			atomic.StoreInt64(&partitionInfo.minOffset, message.Offset)
		}
		return
	}

	partitionInfo = &partitionConsumingInfo{
		consumingMap: make(map[int64]*sarama.ConsumerMessage),
	}
	atomic.StoreInt64(&partitionInfo.minOffset, message.Offset)
	atomic.StoreInt64(&partitionInfo.maxOffset, message.Offset)
	partitionInfo.consumingMap[message.Offset] = message
	c.consumingInfo[message.Partition] = partitionInfo

}

func (c *DataSyncConsumer) process(ctx context.Context, sess sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) {
	onceLog := logger.CtxSugar(ctx)
	start := time.Now()
	defer func() {
		if r := recover(); r != nil {
			// TODO: add monitor report
			errMsg := fmt.Sprintf("err:%+v", r)
			onceLog.Errorf(errMsg)
			c.handleConsumeFail(ctx, msg, r.(error))
			c.finishConsume(ctx, msg)
			c.commitOffset(ctx, sess, msg)
		}
		cost := time.Since(start).Milliseconds()
		// TODO: add monitor report
		onceLog.Infof("message topic: %+v, partition: %+v, offset: %+v, consumed cost: %+v", msg.Topic, msg.Partition, msg.Offset, cost)
	}()
	err := c.KafkaConsumeFunc(ctx, string(msg.Value), msg.Headers)
	if err != nil {
		c.handleConsumeFail(ctx, msg, err)
	}
	c.finishConsume(ctx, msg)
	c.commitOffset(ctx, sess, msg)
}

func (c *DataSyncConsumer) handleConsumeFail(ctx context.Context, msg *sarama.ConsumerMessage, consumeErr error) {
	onceLog := logger.CtxSugar(ctx)
	// TODO: add monitor report
	retryTimes := getConsumeRetryTimes(msg)
	allowRetryTimes := c.RetryTimes
	if retryTimes >= allowRetryTimes {
		// TODO: add monitor report
		onceLog.Errorf("kafka consume message error finally, retryTimes larger than max retries: %+v, err:%+v, msg: %+v, msgValue: %+v", allowRetryTimes, consumeErr, msg, string(msg.Value))
		return
	}
	onceLog.Errorf("kafka consume message error, err:%+v, msg: %+v, msgValue: %+v", consumeErr, msg, string(msg.Value))
	// the time delay strategy is that according to the configuration of DelayTime when retryTimes is less than len(delayTime)
	// otherwise keep delayTime[len(delayTime) -1]
	allowDelayTime := c.DelayTime
	delayTime := allowDelayTime[len(allowDelayTime)-1]
	if int(retryTimes) < len(allowDelayTime) {
		delayTime = allowDelayTime[retryTimes]
	}
	for i := 0; i < 3; i++ {
		addRetryErr := addConsumeRetryTimes(msg)
		if addRetryErr != nil {
			onceLog.Errorf("kafka consume message add retry times err: %+v, msg: %+v, msgValue: %+v", addRetryErr, msg, string(msg.Value))
			return
		}
		sendErr := produce.SendKafkaConsumeMessage(ctx, msg, delayTime)
		if sendErr == nil {
			return
		}
		onceLog.Errorf("kafka consume message send retry err: %+v, msg: %+v", sendErr, msg)
	}
	// TODO: add monitor report
	onceLog.Errorf("kafka consume message send retry failed, msg: %+v", msg)
}

func (c *DataSyncConsumer) finishConsume(ctx context.Context, msg *sarama.ConsumerMessage) {
	onceLog := logger.CtxSugar(ctx)
	<-c.chanMap[msg.Partition]
	c.Lock()
	defer c.Unlock()
	partitionInfo, ok := c.consumingInfo[msg.Partition]
	if !ok {
		onceLog.Warnf("kafka consumer can't find the consume info, topic: %+v, partition: %+v, consume group: %+v", msg.Topic, msg.Partition, c.GroupId)
		return
	}
	// delete partitionInfo.ConsumingMap already processed the message
	delete(partitionInfo.consumingMap, msg.Offset)
	if msg.Offset > partitionInfo.minOffset {
		// it indicates that consumers have processed earlier messages in their previous consumption process, partitionInfo.minOffset < msg.Offset so no need to update
		onceLog.Infof("consume currentOffset is larger than minOffset, current: %+v, minOffset: %+v", msg.Offset, partitionInfo.minOffset)
		return
	}
	// minOffset represents the offset of the earliest message in that partition that is not consumed
	// if msg.Offset <= partitionInfo.minOffset, update partitionInfo.minOffset
	atomic.StoreInt64(&partitionInfo.minOffset, partitionInfo.maxOffset+1)
	for consumingOffset := range partitionInfo.consumingMap {
		if consumingOffset < partitionInfo.minOffset {
			atomic.StoreInt64(&partitionInfo.minOffset, consumingOffset)
		}
	}

	onceLog.Infof("the minOffset change to: %+v, topic: %+v, partition: %+v, consume group: %+v", partitionInfo.minOffset, msg.Topic, msg.Partition, c.GroupId)
}

func (c *DataSyncConsumer) commitOffset(ctx context.Context, sess sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) {
	onceLog := logger.CtxSugar(ctx)
	c.Lock()
	defer c.Unlock()
	partitionInfo, ok := c.consumingInfo[msg.Partition]
	if !ok {
		// mark this message has been consumed
		sess.MarkMessage(msg, "")
		onceLog.Infof("commit currentOffset: %+v", msg.Offset)
		return
	}
	onceLog.Infof("minOffset: %+v, currentOffset: %+v, topic: %+v, partition: %+v, consume group: %+v", partitionInfo.minOffset, msg.Offset, msg.Topic, msg.Partition, c.GroupId)
	if partitionInfo.minOffset > msg.Offset {
		// update the offset(the next offset of the processed message) of this consumer group in this partition
		sess.MarkOffset(msg.Topic, msg.Partition, partitionInfo.minOffset, "")
		onceLog.Infof("commit minOffset: %+v", partitionInfo.minOffset)
	}
}

func generateMsgCtx(msg *sarama.ConsumerMessage) context.Context {
	traceId := getStrFromMsgHeader(msg, constant.KafkaHeaderKeyTraceId)
	ctx := logger.NewTraceIdLog(traceId)
	return ctx
}

func getStrFromMsgHeader(msg *sarama.ConsumerMessage, key string) string {
	for _, header := range msg.Headers {
		if string(header.Key) == key {
			return string(header.Value)
		}
	}
	return ""
}

func getConsumeRetryTimes(message *sarama.ConsumerMessage) uint32 {
	retryStr := getStrFromMsgHeader(message, constant.KafkaHeaderKeyRetryTimes)
	retryTimes, err := strconv.Atoi(retryStr)
	if err != nil {
		return 0
	}
	return uint32(retryTimes)
}

func addConsumeRetryTimes(message *sarama.ConsumerMessage) error {
	for _, header := range message.Headers {
		if string(header.Key) == constant.KafkaHeaderKeyRetryTimes {
			retryTimes, err := strconv.Atoi(string(header.Value))
			if err != nil {
				return err
			}
			header.Value = []byte(strconv.Itoa(retryTimes + 1))
			return nil
		}
	}
	recordHeader := new(sarama.RecordHeader)
	recordHeader.Key = []byte(constant.KafkaHeaderKeyRetryTimes)
	recordHeader.Value = []byte(strconv.Itoa(1))
	message.Headers = append(message.Headers, recordHeader)
	return nil
}
