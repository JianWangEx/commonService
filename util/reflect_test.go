// Package util @Author  wangjian    2023/8/13 1:26 PM
package util

import (
	"fmt"
	"reflect"
	"testing"
)

func TestReflectGrammar(t *testing.T) {
	var arr []*string
	var val *string

	rv1 := reflect.ValueOf(&arr)
	rv2 := reflect.ValueOf(val)
	if rv1.Kind() != reflect.Ptr || rv1.IsNil() {
		fmt.Printf("rv1 is a nil pointer")
	}
	if rv2.Kind() != reflect.Ptr || rv2.IsNil() {
		fmt.Printf("rv2 is a nil pointer")
	}
}
