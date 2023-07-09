// Package util @Author  wangjian    2023/6/22 11:59 PM
package util

import (
	"fmt"
	"strconv"
)

func StringToPtr(str string) *string {
	return &str
}

func convertStringToBasicDataType(s, t string) (interface{}, error) {
	// 根据 typeName 进行类型断言
	switch t {
	case "int8":
		val, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, err
		}
		return int8(val), nil
	case "int":
		val, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, err
		}
		return int(val), nil
	case "int16":
		val, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, err
		}
		return int16(val), nil
	case "int32":
		val, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, err
		}
		return int32(val), nil
	case "int64":
		val, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, err
		}
		return int64(val), nil
	case "uint8":
		val, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, err
		}
		return uint8(val), nil
	case "uint":
		val, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, err
		}
		return uint(val), nil
	case "uint16":
		val, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, err
		}
		return uint16(val), nil
	case "uint32":
		val, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, err
		}
		return uint32(val), nil
	case "uint64":
		val, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, err
		}
		return uint64(val), nil
	case "bytes":
		return []byte(s), nil
	case "float32":
		val, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, err
		}
		return float32(val), nil
	case "float64":
		val, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, err
		}
		return float64(val), nil
	case "string":
		return s, nil
	case "bool":
		return strconv.ParseBool(s)
	default:
		return nil, fmt.Errorf("Unknown type: %v", t)
	}
}
