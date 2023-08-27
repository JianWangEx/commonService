// Package util @Author  wangjian    2023/6/22 10:32 PM
package util

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
)

// DeductPointerVal 返回最终非指针类型的值
func DeductPointerVal(val interface{}) reflect.Value {
	v := reflect.ValueOf(val)
	for v.Kind() == reflect.Ptr {
		if v.IsValid() {
			v = v.Elem()
		} else {
			break
		}
	}
	return v
}

// SetValue uses reflection to set underlying value to receiver
func SetValue(val interface{}, receiver interface{}) error {
	receiverVal := DeductPointerVal(receiver)
	if receiverVal.CanSet() {
		if reflect.TypeOf(val).AssignableTo(receiverVal.Type()) {
			receiverVal.Set(reflect.ValueOf(val))
		} else {
			return errors.New("value cannot assign")
		}
	} else {
		receiverVal.Set(reflect.Zero(receiverVal.Type()))
	}
	return nil
}

// GetValue uses reflection to get underlying value
func GetValue(value interface{}) interface{} {
	var valToStore interface{}
	v := DeductPointerVal(value)
	if v.IsValid() {
		valToStore = v.Interface()
	}

	return valToStore
}

func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}

	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

// IsPointerPointToNil judge whether pointer is point to a nil
func IsPointerPointToNil(i interface{}) bool {
	for true {
		if i == nil {
			return true
		}
		if reflect.ValueOf(i).Kind() == reflect.Ptr {
			if reflect.ValueOf(i).IsNil() {
				return true
			}
			i = reflect.ValueOf(i).Elem().Interface()
		} else {
			break
		}
	}
	return false
}

// AssertValue
//
//	@Description: 根据typeName断言value至指定类型
//	@param value
//	@param typeName
//	@return interface{}
//	@return error
func AssertValue(value interface{}, typeName string) (interface{}, error) {
	// 获取 value 的反射值
	refValue := reflect.ValueOf(value)

	// 根据 typeName 进行类型断言
	switch typeName {
	case "int":
		if refValue.Kind() == reflect.Int {
			return refValue.Int(), nil
		}
	case "int8":
		if refValue.Kind() == reflect.Int8 {
			return refValue.Int(), nil
		}
	case "int16":
		if refValue.Kind() == reflect.Int16 {
			return refValue.Int(), nil
		}
	case "int32", "rune":
		if refValue.Kind() == reflect.Int32 {
			return refValue.Int(), nil
		}
	case "int64":
		if refValue.Kind() == reflect.Int64 {
			return refValue.Int(), nil
		}
	case "uint":
		if refValue.Kind() == reflect.Uint {
			return refValue.Uint(), nil
		}
	case "uint8", "byte":
		if refValue.Kind() == reflect.Uint8 {
			return refValue.Uint(), nil
		}
	case "uint16":
		if refValue.Kind() == reflect.Uint16 {
			return refValue.Uint(), nil
		}
	case "uint32":
		if refValue.Kind() == reflect.Uint32 {
			return refValue.Uint(), nil
		}
	case "uint64":
		if refValue.Kind() == reflect.Uint64 {
			return refValue.Uint(), nil
		}
	case "float32":
		if refValue.Kind() == reflect.Float32 {
			return refValue.Float(), nil
		}
	case "float64":
		if refValue.Kind() == reflect.Float64 {
			return refValue.Float(), nil
		}
	case "string":
		if refValue.Kind() == reflect.String {
			return refValue.String(), nil
		}
	case "bool":
		if refValue.Kind() == reflect.Bool {
			return refValue.Bool(), nil
		}
	default:
		return nil, fmt.Errorf("Unknown type: %v", typeName)
	}

	return nil, fmt.Errorf("Type assertion failed: %v", typeName)
}

// IsZeroOfUnderlyingType return the param x is zero value or not
// Specially, if x is pointer, only it is nil will return true.
// Because zero value of a pointer is nil.
func IsZeroOfUnderlyingType(x interface{}) bool {
	if x == nil {
		return true
	}
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}
