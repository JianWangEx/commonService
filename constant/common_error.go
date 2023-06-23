// Package constant @Author  wangjian    2023/6/23 1:50 PM
package constant

import "fmt"

type CommonError struct {
	name         string
	msg          string
	code         int64
	subErrorCode int64
}

func (e *CommonError) Error() string {
	if e.subErrorCode > 0 {
		return fmt.Sprintf("error_code=%d,sub_error_code=%d,msg=%s", e.code, e.subErrorCode, e.msg)
	}
	return fmt.Sprintf("error_code=%d,msg=%s", e.code, e.msg)
}

var (
	CommonErrorServer = newError("ERROR_SERVER", "ERROR_SERVER", ErrorServer, 0)
)

func newError(name, msg string, code, subErrorCode int64) *CommonError {
	return &CommonError{
		name:         name,
		msg:          msg,
		code:         code,
		subErrorCode: subErrorCode,
	}
}

func (e *CommonError) WithMsgF(format string, a ...interface{}) *CommonError {
	return e.WithMsg(fmt.Sprintf(format, a...))
}

func (e *CommonError) WithMsg(msg string) *CommonError {
	return &CommonError{
		name:         e.name,
		msg:          msg,
		code:         e.code,
		subErrorCode: e.subErrorCode,
	}
}
