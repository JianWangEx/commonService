// Package delay @Author  wangjian    2023/8/2 6:26 PM
package delay

import (
	"github.com/JianWangEx/commonService/constant"
)

var (
	delayTopicMap = make(map[uint32]string)
	delayTimes    = constant.DelayTimes
)

func GetDelayTopic(delayTime uint32, topic string) string {
	if delayTime == 0 || len(delayTopicMap) == 0 {
		return topic
	}

	queueDelayTime := getQueueDelayTime(delayTime)
	return delayTopicMap[queueDelayTime]
}

func getQueueDelayTime(delayTime uint32) uint32 {
	queueDelayTime := uint32(0)
	for _, v := range delayTimes {
		queueDelayTime = v
		if queueDelayTime >= delayTime {
			break
		}
	}
	return queueDelayTime
}
