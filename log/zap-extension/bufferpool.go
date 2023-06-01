// Package zap_extension @Author  wangjian    2023/6/1 6:18 PM
package zap_extension

import "go.uber.org/zap/buffer"

var (
	_bufferPool = buffer.NewPool()

	// getBuffer 从池中检索一个buffer，必要时创建一个
	getBuffer = _bufferPool.Get
)
