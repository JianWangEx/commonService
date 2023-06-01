// Package zap_extension @Author  wangjian    2023/6/1 5:51 PM
package zap_extension

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"sync"
)

// _consolePool 同步对象池(sync.Pool), 用于复用consoleEncoder对象
var _consolePool = sync.Pool{
	New: func() interface{} {
		return consoleEncoder{}
	},
}

// _sliceEncoderPool 同步对象池(sync.Pool), 用于复用sliceArrayEncoder对象
var _sliceEncoderPool = sync.Pool{
	New: func() interface{} {
		return &sliceArrayEncoder{elems: make([]interface{}, 0, 2)}
	},
}

type consoleEncoder struct {
	*EncoderConfig
	buf            *buffer.Buffer
	spaced         bool
	openNamespaces int
	traceID        string

	// for encoding generic values by reflection
	reflectBuf *buffer.Buffer
	reflectEnc *json.Encoder
}

// Clone 复制当前encoder，确保对复制体添加字段不会影响原始encoder。该方法返回一个新的encoder对象。
func (c *consoleEncoder) Clone() zapcore.Encoder {
	clone := c.clone()
	clone.buf.Write(c.buf.Bytes())
	return clone
}

// EncodeEntry 将一个日志条目（Entry）以及与之关联的字段（Field）以及任何积累的context编码为byte buffer，
// 并返回结果。 任何空的字段，包括 Entry 类型的字段，都应该被省略。
func (c *consoleEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	final := c.clone()
	defer func() {
		putConsoleEncoder(final)
	}()

	line := getBuffer()

	// 我们不希望将日志entry的元数据进行引号和转义处理（如果它们被编码为字符串）。
	// 因此，不能使用 JSON 编码器对日志entry进行编码。
	// 为了简单起见，选择使用内存编码器（memory encoder）和 fmt.Fprint 函数
	//
	// 如果这种实现方式在性能上成为瓶颈，可以考虑实现针对纯文本格式的 ArrayEncoder 接口
	arr := getSliceEncoder()
	if final.TimeKey != "" && final.EncodeTime != nil {
		final.EncodeTime(c.Time, arr)
	}
	if final.LevelKey != "" && final.EncodeLevel != nil {
		final.EncodeLevel(c.Level, arr)
	}
	if c.LoggerName != "" && final.NameKey != "" {
		nameEncoder := final.EncodeName
		if nameEncoder == nil {
			// 回退到FullNameEncoder 已实现向后兼容性
			nameEncoder = zapcore.FullNameEncoder
		}

		nameEncoder(c.LoggerName, arr)
	}
	if c.Caller.Defined {
		if final.CallerKey != "" && final.EncodeCaller != nil {
			final.EncodeCaller(ent.Caller, arr)
		}
	}

	// 添加trace id
	if final.traceID != "" {
		arr.AppendString(final.traceID)
	} else {
		arr.AppendString("")
	}

	for i := range arr.elems {
		if i > 0 {
			line.AppendString(c.ConsoleSeparator)
		}
		fmt.Fprint(line, arr.elems[i])
	}
	putSliceEncoder(arr)
}

func (c *consoleEncoder) clone() *consoleEncoder {
	clone := getConsoleEncoder()
	clone.TraceKey = c.traceID
	clone.EncoderConfig = c.EncoderConfig
	clone.openNamespaces = c.openNamespaces
	clone.buf = getBuffer()
	return clone
}

// 从对象池中获取
func getConsoleEncoder() *consoleEncoder {
	return _consolePool.Get().(*consoleEncoder)
}

// 放回对象池中
func putConsoleEncoder(c *consoleEncoder) {
	if c.reflectBuf != nil {
		c.reflectBuf.Free()
	}
	c.EncoderConfig = nil
	c.buf = nil
	c.openNamespaces = nil
	c.reflectBuf = nil
	c.reflectEnc = nil
	_consolePool.Put(c)
}

func getSliceEncoder() *sliceArrayEncoder {
	return _sliceEncoderPool.Get().(*sliceArrayEncoder)
}

func putSliceEncoder(s *sliceArrayEncoder) {
	s.elems = s.elems[:0]
	_sliceEncoderPool.Put(s)
}

// NewConsoleEncoder 创建一个encoder，其输出是为人类而不是及其消费而设计的。它以纯文本格式序列化core日志条目数据
// （message，level，timestamp，etc）并将结构化上下文保留为JSON
//
// 注意，尽管console encoder不使用encoder配置中指定的key，但它将忽略key设置为空字符串的任何元素
func NewConsoleEncoder(encConfig EncoderConfig) zapcore.Encoder {
	if encConfig.ConsoleSeparator == "" {
		// 如果控制台分隔符为空，使用默认"\t"实现向后兼容
		encConfig.ConsoleSeparator = "\t"
	}
	return &consoleEncoder{
		EncoderConfig: &encConfig,
		buf:           getBuffer(),
		spaced:        true,
	}
}
