// Package zap_extension @Author  wangjian    2023/6/1 5:51 PM
package zap_extension

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"math"
	"sync"
	"time"
	"unicode/utf8"
)

// For JSON-escaping; see jsonEncoder.safeAddString below.
const _hex = "0123456789abcdef"

// _consolePool 同步对象池(sync.Pool), 用于复用consoleEncoder对象
var _consolePool = sync.Pool{
	New: func() interface{} {
		return &consoleEncoder{}
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
		final.EncodeTime(ent.Time, arr)
	}
	if final.LevelKey != "" && final.EncodeLevel != nil {
		final.EncodeLevel(ent.Level, arr)
	}
	if ent.LoggerName != "" && final.NameKey != "" {
		nameEncoder := final.EncodeName
		if nameEncoder == nil {
			// 回退到FullNameEncoder 已实现向后兼容性
			nameEncoder = zapcore.FullNameEncoder
		}

		nameEncoder(ent.LoggerName, arr)
	}
	if ent.Caller.Defined {
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

	// 添加message
	if final.MessageKey != "" {
		final.addSeparatorIfNecessary(line)
		line.AppendString(ent.Message)
	}

	if c.buf.Len() > 0 {
		final.addElementSeparator()
		final.buf.Write(c.buf.Bytes())
	}
	// 添加任何结构化context
	final.writeContext(line, fields)

	// 如果没有stacktrace key，尊重它，这允许用户强制单行的输出
	if ent.Stack != "" && c.StacktraceKey != "" {
		line.AppendByte('\n')
		line.AppendString(ent.Stack)
	}

	if final.LineEnding != "" {
		line.AppendString(c.LineEnding)
	} else {
		line.AppendString(zapcore.DefaultLineEnding)
	}

	return line, nil
}

func (c *consoleEncoder) AddInt(k string, v int)         { c.AddInt64(k, int64(v)) }
func (c *consoleEncoder) AddInt32(k string, v int32)     { c.AddInt64(k, int64(v)) }
func (c *consoleEncoder) AddInt16(k string, v int16)     { c.AddInt64(k, int64(v)) }
func (c *consoleEncoder) AddInt8(k string, v int8)       { c.AddInt64(k, int64(v)) }
func (c *consoleEncoder) AddUint(k string, v uint)       { c.AddUint64(k, uint64(v)) }
func (c *consoleEncoder) AddUint32(k string, v uint32)   { c.AddUint64(k, uint64(v)) }
func (c *consoleEncoder) AddUint16(k string, v uint16)   { c.AddUint64(k, uint64(v)) }
func (c *consoleEncoder) AddUint8(k string, v uint8)     { c.AddUint64(k, uint64(v)) }
func (c *consoleEncoder) AddUintptr(k string, v uintptr) { c.AddUint64(k, uint64(v)) }
func (c *consoleEncoder) AppendComplex64(v complex64)    { c.AppendComplex128(complex128(v)) }
func (c *consoleEncoder) AppendComplex128(v complex128) {
	c.addElementSeparator()
	// 转换成独立于平台，固定大小类型
	// lint: unnecessary conversion (unconvert).
	r, i := real(v), imag(v)
	c.buf.AppendByte('"')
	// 因为我们总是在一个带引号的字符串中，所以我们可以使用strconv
	// 而不使用特殊大小写的NaN和+/-Inf
	c.buf.AppendFloat(r, 64)
	c.buf.AppendByte('+')
	c.buf.AppendFloat(i, 64)
	c.buf.AppendByte('i')
	c.buf.AppendByte('"')
}
func (c *consoleEncoder) AppendFloat64(v float64)            { c.appendFloat(v, 64) }
func (c *consoleEncoder) AppendFloat32(v float32)            { c.appendFloat(float64(v), 32) }
func (c *consoleEncoder) AppendInt(v int)                    { c.AppendInt64(int64(v)) }
func (c *consoleEncoder) AppendInt32(v int32)                { c.AppendInt64(int64(v)) }
func (c *consoleEncoder) AppendInt16(v int16)                { c.AppendInt64(int64(v)) }
func (c *consoleEncoder) AppendInt8(v int8)                  { c.AppendInt64(int64(v)) }
func (c *consoleEncoder) AppendUint(v uint)                  { c.AppendUint64(uint64(v)) }
func (c *consoleEncoder) AppendUint32(v uint32)              { c.AppendUint64(uint64(v)) }
func (c *consoleEncoder) AppendUint16(v uint16)              { c.AppendUint64(uint64(v)) }
func (c *consoleEncoder) AppendUint8(v uint8)                { c.AppendUint64(uint64(v)) }
func (c *consoleEncoder) AppendUintptr(v uintptr)            { c.AppendUint64(uint64(v)) }
func (c *consoleEncoder) AddComplex64(k string, v complex64) { c.AddComplex128(k, complex128(v)) }
func (c *consoleEncoder) AddFloat32(k string, v float32)     { c.AddFloat64(k, float64(v)) }

func (c *consoleEncoder) AppendUint64(val uint64) {
	c.addElementSeparator()
	c.buf.AppendUint(val)
}

func (c *consoleEncoder) AppendTime(val time.Time) {
	cur := c.buf.Len()
	if e := c.EncodeTime; e != nil {
		e(val, c)
	}
	if cur == c.buf.Len() {
		// User-supplied EncodeTime is a no-op. Fall back to nanos since epoch to keep
		// output JSON valid.
		c.AppendInt64(val.UnixNano())
	}
}

func (c *consoleEncoder) AppendString(val string) {
	c.addElementSeparator()
	c.buf.AppendByte('"')
	c.safeAddString(val)
	c.buf.AppendByte('"')
}

func (c *consoleEncoder) AppendReflected(val interface{}) error {
	valueBytes, err := c.encodeReflected(val)
	if err != nil {
		return err
	}
	c.addElementSeparator()
	_, err = c.buf.Write(valueBytes)
	return err
}

func (c *consoleEncoder) AppendInt64(val int64) {
	c.addElementSeparator()
	c.buf.AppendInt(val)
}

func (c *consoleEncoder) AppendDuration(val time.Duration) {
	cur := c.buf.Len()
	if e := c.EncodeDuration; e != nil {
		e(val, c)
	}
	if cur == c.buf.Len() {
		// User-supplied EncodeDuration is a no-op. Fall back to nanoseconds to keep
		// JSON valid.
		c.AppendInt64(int64(val))
	}
}

func (c *consoleEncoder) clone() *consoleEncoder {
	clone := getConsoleEncoder()
	clone.TraceKey = c.traceID
	clone.EncoderConfig = c.EncoderConfig
	clone.openNamespaces = c.openNamespaces
	clone.buf = getBuffer()
	return clone
}

func (c *consoleEncoder) AppendByteString(val []byte) {
	c.addElementSeparator()
	c.buf.AppendByte('"')
	c.safeAddByteString(val)
	c.buf.AppendByte('"')
}

func (c *consoleEncoder) AppendBool(val bool) {
	c.addElementSeparator()
	c.buf.AppendBool(val)
}

func (c *consoleEncoder) AppendObject(obj zapcore.ObjectMarshaler) error {
	// Close ONLY new openNamespaces that are created during
	// AppendObject().
	old := c.openNamespaces
	c.openNamespaces = 0
	c.addElementSeparator()
	c.buf.AppendByte('{')
	err := obj.MarshalLogObject(c)
	c.buf.AppendByte('}')
	c.closeOpenNamespaces()
	c.openNamespaces = old
	return err
}

func (c *consoleEncoder) AppendArray(arr zapcore.ArrayMarshaler) error {
	c.addElementSeparator()
	c.buf.AppendByte('[')
	err := arr.MarshalLogArray(c)
	c.buf.AppendByte(']')
	return err
}

func (c *consoleEncoder) AddUint64(key string, val uint64) {
	c.addKey(key)
	c.AppendUint64(val)
}

func (c *consoleEncoder) AddTime(key string, val time.Time) {
	c.addKey(key)
	c.AppendTime(val)
}

func (c *consoleEncoder) AddString(key, val string) {
	switch key {
	case TraceKey:
		c.traceID = val
	default:
		c.addKey(key)
		c.AppendString(val)
	}
}

func (c *consoleEncoder) OpenNamespace(key string) {
	c.addKey(key)
	c.buf.AppendByte('{')
	c.openNamespaces++
}

func (c *consoleEncoder) AddReflected(key string, obj interface{}) error {
	valueBytes, err := c.encodeReflected(obj)
	if err != nil {
		return err
	}
	c.addKey(key)
	_, err = c.buf.Write(valueBytes)
	return err
}

var nullLiteralBytes = []byte("null")

// Only invoke the standard JSON encoder if there is actually something to
// encode; otherwise write JSON null literal directly.
func (c *consoleEncoder) encodeReflected(obj interface{}) ([]byte, error) {
	if obj == nil {
		return nullLiteralBytes, nil
	}
	c.resetReflectBuf()
	if err := c.reflectEnc.Encode(obj); err != nil {
		return nil, err
	}
	c.reflectBuf.TrimNewline()
	return c.reflectBuf.Bytes(), nil
}

func (c *consoleEncoder) resetReflectBuf() {
	if c.reflectBuf == nil {
		c.reflectBuf = getBuffer()
		c.reflectEnc = json.NewEncoder(c.reflectBuf)

		// 为了与我们自定义encoder保持一致
		c.reflectEnc.SetEscapeHTML(false)
	} else {
		c.reflectBuf.Reset()
	}
}

func (c *consoleEncoder) AddInt64(key string, val int64) {
	c.addKey(key)
	c.AppendInt64(val)
}

func (c *consoleEncoder) AddFloat64(key string, val float64) {
	c.addKey(key)
	c.AppendFloat64(val)
}

func (c *consoleEncoder) AddDuration(key string, val time.Duration) {
	c.addKey(key)
	c.AppendDuration(val)
}

func (c *consoleEncoder) AddArray(key string, arr zapcore.ArrayMarshaler) error {
	c.addKey(key)
	return c.AppendArray(arr)
}

func (c *consoleEncoder) AddObject(key string, obj zapcore.ObjectMarshaler) error {
	c.addKey(key)
	return c.AppendObject(obj)
}

func (c *consoleEncoder) AddBinary(key string, val []byte) {
	c.AddString(key, base64.StdEncoding.EncodeToString(val))
}

func (c *consoleEncoder) AddByteString(key string, val []byte) {
	c.addKey(key)
	c.AppendByteString(val)
}

func (c *consoleEncoder) AddBool(key string, val bool) {
	c.addKey(key)
	c.AppendBool(val)
}

func (c *consoleEncoder) AddComplex128(key string, val complex128) {
	c.addKey(key)
	c.AppendComplex128(val)
}

func (c *consoleEncoder) addElementSeparator() {
	last := c.buf.Len() - 1
	if last < 0 {
		return
	}
	switch c.buf.Bytes()[last] {
	case '{', '[', ':', ',', ' ':
		return
	default:
		c.buf.AppendByte(',')
		if c.spaced {
			c.buf.AppendByte(' ')
		}
	}
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
	c.openNamespaces = 0
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

// NewConsoleEncoder 创建一个encoder，其输出是为人类而不是计算机消费而设计的。它以纯文本格式序列化core日志条目数据
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

func (c *consoleEncoder) closeOpenNamespaces() {
	for i := 0; i < c.openNamespaces; i++ {
		c.buf.AppendByte('}')
	}
	c.openNamespaces = 0
}

func (c *consoleEncoder) addKey(key string) {
	c.addElementSeparator()
	c.buf.AppendByte('"')
	c.safeAddString(key)
	c.buf.AppendByte('"')
	c.buf.AppendByte(':')
	if c.spaced {
		c.buf.AppendByte(' ')
	}
}

func (c *consoleEncoder) appendFloat(val float64, bitSize int) {
	c.addElementSeparator()
	switch {
	case math.IsNaN(val):
		c.buf.AppendString(`"NaN"`)
	case math.IsInf(val, 1):
		c.buf.AppendString(`"+Inf"`)
	case math.IsInf(val, -1):
		c.buf.AppendString(`"-Inf"`)
	default:
		c.buf.AppendFloat(val, bitSize)
	}
}

// safeAddString JSON-escapes a string and appends it to the internal buffer.
// Unlike the standard library's encoder, it doesn't attempt to protect the
// user from browser vulnerabilities or JSONP-related problems.
func (c *consoleEncoder) safeAddString(s string) {
	for i := 0; i < len(s); {
		if c.tryAddRuneSelf(s[i]) {
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if c.tryAddRuneError(r, size) {
			i++
			continue
		}
		c.buf.AppendString(s[i : i+size])
		i += size
	}
}

// safeAddByteString is no-alloc equivalent of safeAddString(string(s)) for s []byte.
func (c *consoleEncoder) safeAddByteString(s []byte) {
	for i := 0; i < len(s); {
		if c.tryAddRuneSelf(s[i]) {
			i++
			continue
		}
		r, size := utf8.DecodeRune(s[i:])
		if c.tryAddRuneError(r, size) {
			i++
			continue
		}
		c.buf.Write(s[i : i+size])
		i += size
	}
}

// tryAddRuneSelf appends b if it is valid UTF-8 character represented in a single byte.
func (c *consoleEncoder) tryAddRuneSelf(b byte) bool {
	if b >= utf8.RuneSelf {
		return false
	}
	if 0x20 <= b && b != '\\' && b != '"' {
		c.buf.AppendByte(b)
		return true
	}
	switch b {
	case '\\', '"':
		c.buf.AppendByte('\\')
		c.buf.AppendByte(b)
	case '\n':
		c.buf.AppendByte('\\')
		c.buf.AppendByte('n')
	case '\r':
		c.buf.AppendByte('\\')
		c.buf.AppendByte('r')
	case '\t':
		c.buf.AppendByte('\\')
		c.buf.AppendByte('t')
	default:
		// Encode bytes < 0x20, except for the escape sequences above.
		c.buf.AppendString(`\u00`)
		c.buf.AppendByte(_hex[b>>4])
		c.buf.AppendByte(_hex[b&0xF])
	}
	return true
}

func (c *consoleEncoder) tryAddRuneError(r rune, size int) bool {
	if r == utf8.RuneError && size == 1 {
		c.buf.AppendString(`\ufffd`)
		return true
	}
	return false
}

func (c *consoleEncoder) addSeparatorIfNecessary(line *buffer.Buffer) {
	if line.Len() > 0 {
		line.AppendString(c.ConsoleSeparator)
	}
}

func addFields(enc zapcore.ObjectEncoder, fields []zapcore.Field) {
	for i := range fields {
		fields[i].AddTo(enc)
	}
}

func (c *consoleEncoder) writeContext(line *buffer.Buffer, extra []zapcore.Field) {
	addFields(c, extra)
	c.closeOpenNamespaces()
	if c.buf.Len() == 0 {
		return
	}

	c.addSeparatorIfNecessary(line)
	line.AppendByte('{')
	line.Write(c.buf.Bytes())
	line.AppendByte('}')
}
