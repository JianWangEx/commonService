// Package constant @Author  wangjian    2023/6/21 8:25 PM
package constant

import "errors"

var (
	// ErrorNilReceiverOrNotPtr means receiver is nil or not a pointer
	ErrorNilReceiverOrNotPtr = errors.New("receiver is nil or not a ptr")
	// ErrorCacheMiss means that a Get failed because the item wasn't present
	ErrorCacheMiss = errors.New("cache miss")
	// ErrorFailedOperation means redis return failed
	ErrorFailedOperation = errors.New("operation failed")
)
