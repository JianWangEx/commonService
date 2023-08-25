// Package cronjob @Author  wangjian    2023/8/26 00:36
package cronjob

import "context"

type RunOnce interface {
	// JobParseParam used to parse the input string to the job request parameters
	JobParseParam(context.Context, string) (interface{}, error)

	// RunJobMethod used to run the job method
	// the input parameters is the JobParseParam's return interface
	RunJobMethod(context.Context, interface{}) error
}

var RunOnceMap = make(map[string]RunOnce)

func RegisterRunOnce(jobName string, method RunOnce) {
	RunOnceMap[jobName] = method
}
