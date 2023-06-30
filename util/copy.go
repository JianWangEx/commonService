// Package util @Author  wangjian    2023/6/29 7:02 PM
package util

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

const (
	errorNilPointer = "src or dst can not be nil"
	errorEncoding   = "encoding failed during deep copy"
	errorDecoding   = "decoding failed during deep copy"
)

// DeepCopy
//
//	@Description: 将src深拷贝到dst，其性能更接近与gobDeepCopy。
//	对于结构体，如果字段是指针，指针的值是默认值，则gobDeepCopy无法复制默认值，所以推荐使用jsoniterDeepCopy
//	@param dst
//	@param src
//	@return error
func DeepCopy(dst interface{}, src interface{}) error {
	return jsoniterDeepCopy(dst, src)
}

// json-iterator/go deep copy
var jsoniterJSON = jsoniter.ConfigFastest

func jsoniterDeepCopy(dst interface{}, src interface{}) error {
	if IsPointerPointToNil(dst) || IsPointerPointToNil(src) {
		return errors.New(errorNilPointer)
	}

	encodedBytes, err := jsoniterJSON.Marshal(src)
	if err != nil {
		return errors.New(errorEncoding)
	}
	if err := jsoniterJSON.Unmarshal(encodedBytes, dst); err != nil {
		return errors.New(errorDecoding)
	}

	return nil
}
