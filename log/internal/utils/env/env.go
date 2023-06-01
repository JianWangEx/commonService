// Package env @Author  wangjian    2023/6/1 12:08 PM
package env

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetEnv() string {
	val, ok := os.LookupEnv("ENV")
	if !ok {
		val = "dev"
	}
	return val
}

func IsLive() bool {
	return GetEnv() == "live"
}

func GetFilePath(logDir, fileName string) string {
	if logDir == "" {
		logDir = "./log"
	}
	path := filepath.Join(logDir, fmt.Sprintf("%s.log", fileName))
	return path
}
