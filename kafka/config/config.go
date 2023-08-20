// Package config @Author  wangjian    2023/7/14 9:47 AM
package config

import (
	"errors"
	"github.com/BurntSushi/toml"
	"github.com/Shopify/sarama"
	"strings"
	"sync"
)

var (
	once               sync.Once
	defaultKafkaConfig *sarama.Config

	// kafkaConfig
	config *kafkaConfig

	// kafka consumer cluster name to KafkaCluster map
	kafkaConsumerClusterMap = make(map[string]Sarama)

	// kafka producer topic to cluster name map
	producerTopicToClusterMap = make(map[string]string)
	// kafka consumer topic to cluster name map
	consumerTopicToClusterMap = make(map[string]string)
)

type kafkaConfig struct {
	Sarama Sarama // default config

	ProducerCluster []KafkaCluster // producer cluster name to the cluster address mapping
	ConsumerCluster []KafkaCluster // consumer cluster name to the cluster address mapping

	ProducerTopics []TopicCluster // producer topic name to the cluster name mapping, it points out that it's produced by some producer cluster
	ConsumerTopics []TopicCluster // consumer topic name to the cluster name mapping, it points out that it's consumed by some consumer cluster

	Consumers []Consumer // consumer config, it points out that the specified consumer config includes the topic, delay time and so on.
}

type Sarama struct {
	Brokers  []string // kafka brokers addresses
	UserName string
	Password string
}

type KafkaCluster struct {
	Name string
	Sarama
}

type Consumer struct {
	Topic string
	// group level for kafka, enum
	GroupLevel string
	// when need retry with different time internal, use this when failed to reconsume message
	DelayTime []uint32
	// the count for retry times
	RetryTimes uint32
	// the number of concurrent consumption
	ConcurrentNums uint32
}

type TopicCluster struct {
	Topic       string
	ClusterName string
}

func InitKafkaConfig() {
	once.Do(func() {
		// 初始化defaultKafkaConfig
		defaultKafkaConfig = sarama.NewConfig()
		defaultKafkaConfig.Producer.Return.Successes = true
	})
}

func GetDefaultKafkaConfig() *sarama.Config {
	return defaultKafkaConfig
}

func initKafkaClusterConfigByToml(path string) error {
	config = &kafkaConfig{}
	_, err := toml.DecodeFile(path, config)
	if err != nil {
		return err
	}

	// init producerTopicToClusterMap
	for _, tc := range config.ProducerTopics {
		producerTopicToClusterMap[tc.Topic] = tc.ClusterName
	}

	// init consumerTopicToClusterMap
	for _, tc := range config.ConsumerTopics {
		consumerTopicToClusterMap[tc.Topic] = tc.ClusterName
	}

	// init kafkaConsumerClusterMap
	for _, c := range config.ConsumerCluster {
		kafkaConsumerClusterMap[c.Name] = c.Sarama
	}

	return nil
}

func InitKafkaClusterConfig(path string) error {
	if config != nil {
		return nil
	}
	s := strings.Split(path, ".")
	suffix := s[len(s)-1]
	// 根据不同配置文件格式进行配置
	switch suffix {
	case "toml":
		return initKafkaClusterConfigByToml(path)
	default:
		return errors.New("invalid config file format")
	}
}

func Kafka() kafkaConfig {
	return *config
}

func GetProducerTopicToClusterMap() map[string]string {
	return producerTopicToClusterMap
}

func GetConsumerTopicToClusterMap() map[string]string {
	return consumerTopicToClusterMap
}

func GetKafkaConsumerClusterMap() map[string]Sarama {
	return kafkaConsumerClusterMap
}
