// Package constant @Author  wangjian    2023/6/23 11:32 AM
package constant

import "fmt"

const (
	ErrorSuccess = 0
	ErrorLock    = 1
	ErrorServer  = 2
)

// ErrorDict 标准错误返回格式，如果success，其他字段应为空
type ErrorDict struct {
	Error     string `json:"error,omitempty"`
	ErrorCode int    `json:"error_code,omitempty"`
	ErrorMsg  string `json:"error_msg,omitempty"`
}

var (
	ErrorCodeSuccess     = NewErrorDict(ErrorSuccess, "", "")
	ErrorCodeErrorLock   = NewErrorDict(ErrorLock, "ERROR_LOCK", "ErrorLock")
	ErrorCodeErrorServer = NewErrorDict(ErrorServer, "ERROR_SERVER", "ERROR_SERVER")
)

func NewErrorDict(code int, name string, msg string) ErrorDict {
	return ErrorDict{name, code, msg}
}

func (e ErrorDict) WithMsg(msg string) *ErrorDict {
	newE := e
	newE.ErrorMsg = msg
	return &newE
}

func (e ErrorDict) WithMsgF(format string, args ...interface{}) *ErrorDict {
	newE := e
	newE.ErrorMsg = fmt.Sprintf(format, args...)
	return &newE
}
