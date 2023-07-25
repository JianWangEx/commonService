// Package util @Author  wangjian    2023/7/22 6:15 PM
package util

import (
	"bytes"
	"encoding/json"
	"strings"
)

func SafeToJson(o interface{}) string {
	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	err := jsonEncoder.Encode(o)
	if err != nil {
		return ""
	}
	return strings.TrimRight(bf.String(), "\n")
}
