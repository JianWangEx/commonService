// Package util @Author  wangjian    2023/6/22 10:32 PM
package util

import (
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
