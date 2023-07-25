// Package config @Author  wangjian    2023/7/14 9:47 AM
package config

import (
	"errors"
	"github.com/BurntSushi/toml"
	logger "github.com/JianWangEx/commonService/log"
	"github.com/Shopify/sarama"
	"strings"
	"sync"
)

var (
	once               sync.Once
	defaultKafkaConfig *sarama.Config

	kafkaClusterConfigMap = make(map[string]KafkaCluster)
)

type Sarama struct {
	Brokers  []string
	UserName string
	Password string
}

type KafkaCluster struct {
	Name string
	Sarama
}

type Clusters struct {
	KafkaCluster []KafkaCluster `json:"kafkaCluster"`
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
	clusters := &Clusters{}
	_, err := toml.DecodeFile(path, clusters)
	if err != nil {
		return err
	}

	// 初始化kafkaClusterConfigMap
	for _, c := range clusters.KafkaCluster {
		_, ok := kafkaClusterConfigMap[c.Name]
		if !ok {
			kafkaClusterConfigMap[c.Name] = c
			continue
		}
		logger.GetLogger().Sugar().Warnf("[InitKafkaClusterConfigByToml]duplicate kafka cluster configuration|config path: %s, clusterName: %s", path, c.Name)
	}
	return nil
}

func InitKafkaClusterConfig(path string) error {
	if len(kafkaClusterConfigMap) > 0 {
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

func GetKafkaClusterConfigMap() map[string]KafkaCluster {
	return kafkaClusterConfigMap
}
