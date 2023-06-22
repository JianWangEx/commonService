// Package util @Author  wangjian    2023/6/21 4:36 PM
package util

import (
	"fmt"
	"github.com/BurntSushi/toml"
)

func ParseTomlConfig(path string, config interface{}) error {
	if _, err := toml.DecodeFile(path, config); err != nil {
		return fmt.Errorf("decode file error:%s, path=%s", err, path)
	}
	return nil
}
