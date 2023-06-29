// Package cache @Author  wangjian    2023/6/29 12:19 PM
package cache

import (
	"fmt"
	"testing"
)

type TestFunction interface {
	GetName() string
	GetPhone() string
}

type BaseHandler struct {
}

func (f *BaseHandler) GetName() string {
	fmt.Println("BaseHandler")
	return "BaseHandler"
}

func (f *BaseHandler) GetPhone() string {
	fmt.Println("BaseHandlerPhone")
	return "BaseHandlerPhone"
}

type Student struct {
	Handler BaseHandler
	Name    string
}

func (s *Student) GetName() string {
	fmt.Println(s.Name)
	return s.Name
}

func TestInterface(t *testing.T) {
	student := Student{
		Name: "cat",
	}
	student.GetName()
	student.Handler.GetPhone()
}
