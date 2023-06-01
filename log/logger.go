// Package log @Author  wangjian    2023/6/1 11:03 AM
package log

import (
	"commonService/log/internal/utils/env"
	zaplib "commonService/log/zap-extension"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/hashicorp/go-multierror"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type LogLevel = zapcore.Level
type PrintToStd uint8
type SplitLevel string

const (
	DebugLvl  = zapcore.DebugLevel
	InfoLvl   = zapcore.InfoLevel
	WarnLvl   = zapcore.WarnLevel
	ErrorLvl  = zapcore.ErrorLevel
	PanicLvl  = zapcore.PanicLevel
	DPanicLvl = zapcore.DPanicLevel
	FatalLvl  = zapcore.FatalLevel

	PrintToStd_NONE    PrintToStd = 0
	PrintToStd_USERLOG PrintToStd = 1
	PrintToStd_SYSLOG  PrintToStd = 2
	PrintToStd_TRACING PrintToStd = 4
	PrintToStd_ALL     PrintToStd = 7

	SplitError SplitLevel = "error"
	SplitWarn  SplitLevel = "warn"
	SplitInfo  SplitLevel = "info"
	SplitDebug SplitLevel = "debug"
	SplitNone  SplitLevel = "none"
)

const (
	maxResetLvlDuration = 48 * time.Hour
	maxResetVersion     = 10000

	customTimeLayout = "2006-01-02 15:04:05.999999-07:00"

	DefaultLogFileName     = "server"
	SysErrorLogFileName    = "sys_error"
	SysLogFileName         = "sys"
	DefaultTracingFileName = "traffic_recording"
)

var (
	logger        *zap.Logger
	sysLogger     *zap.Logger
	tracingLogger *zap.Logger

	loggerInitOnce        sync.Once
	sysLoggerInitOnce     sync.Once
	tracingLoggerInitOnce sync.Once

	initialLogLevel LogLevel
	logLevel        atomic.Int32

	resetVer atomic.Int64

	nameMap  = map[LogLevel]string{DebugLvl: "debug", InfoLvl: "info", WarnLvl: "warn", ErrorLvl: "error"}
	levelMap = map[SplitLevel]LogLevel{SplitDebug: DebugLvl, SplitInfo: InfoLvl, SplitWarn: WarnLvl, SplitError: ErrorLvl}
)

type Config struct {
	// Level 初始化默认日志级别
	// 默认情况下，DebugLvl被设置为非线上环境，InfoLvl被设置为线上环境
	Level LogLevel

	// PrintToStd 输出哪一种日志到stdout，默认为空，仅影响非线上环境
	PrintToStd PrintToStd

	// Compress 是否压缩
	Compress bool

	// Path 自定义日志文件路径。仅在K8S中有效。如果没有指定，日志文件将创建在./Log目录下
	Path string

	// LogFileName 自定义日志文件名，如果没有指定，它将是server.log
	LogFileName string

	// SplitLevel 日志分裂的最小等级，大于此级别的日志将会被写入不同的文件，小于此级别的日志被写入server.log，如果不设置所有日志都写入server.log
	SplitLevel SplitLevel

	// TracingLogFileName 自定义跟踪日志文件，如果没有指定，它将是traffic_recording.log
	TracingLogFileName string
}

// InitLogger 初始化logger和system logger
// 这个function应该只运行一次
func InitLogger(config *Config) {
	// 初始化log level
	initLogLevel(config)

	loggerInitOnce.Do(func() {
		// 初始化默认logger
		initDefaultLogger(config)
	})

	sysLoggerInitOnce.Do(func() {
		// 初始化system logger
		initSystemLogger(config)
	})

	tracingLoggerInitOnce.Do(func() {
		// 初始化 tracing logger
		initTracingLogger(config)
	})

}

func initLogLevel(config *Config) {
	level := defaultLevel()
	if config.Level > level {
		level = config.Level
	}
	initialLogLevel = level
	SetLevel(level, 0)

}

// GetLogger 返回logger，log将会输出到./log/error.file和./log/server.log
func GetLogger() *zap.Logger {
	loggerInitOnce.Do(func() {
		config := &Config{
			Level:      InfoLvl,
			SplitLevel: SplitNone,
			PrintToStd: PrintToStd_NONE,
		}
		initDefaultLogger(config)
	})
	return logger
}

// GetSysLogger 返回system logger
// 日志将输出到./log/sys_error.log和./log/sys.log
// 这个logger应只被用于系统框架，请使用GetLogger()对于业务日志
func GetSysLogger() *zap.Logger {
	sysLoggerInitOnce.Do(func() {
		config := &Config{
			Level:      InfoLvl,
			PrintToStd: PrintToStd_NONE,
		}
		initSystemLogger(config)
	})
	return sysLogger
}

func GetTracingLogger() *zap.Logger {
	tracingLoggerInitOnce.Do(func() {
		config := &Config{
			Level:              InfoLvl,
			TracingLogFileName: DefaultTracingFileName,
		}
		initTracingLogger(config)
	})
	return tracingLogger
}

func Sync() error {
	var res *multierror.Error
	if err := GetLogger().Sync(); err != nil {
		res = multierror.Append(res, err)
	}
	if err := GetSysLogger().Sync(); err != nil {
		res = multierror.Append(res, err)
	}
	if err := GetTracingLogger().Sync(); err != nil {
		res = multierror.Append(res, err)
	}
	return res
}

func defaultLevel() zapcore.Level {
	if env.IsLive() {
		return zapcore.InfoLevel
	}
	return zapcore.DebugLevel
}

// SetLevel 动态设置日志级别，duration只适用于在live环境下设置日志级别为debug时，
// 即在live环境下设置日志级别为debug时，在设置时间间隔之后，日志级别将恢复为初始配置（如果为设置，默认为InfoLvl）
func SetLevel(level zapcore.Level, duration time.Duration) {
	if env.IsLive() && level < zapcore.InfoLevel {
		if duration > maxResetLvlDuration {
			duration = maxResetLvlDuration
		}

		ver := resetVer.Add(1)
		if ver > maxResetVersion {
			ver = ver % maxResetVersion
			resetVer.Store(ver)
		}

		time.AfterFunc(duration, func() {
			resetLevel(ver)
		})
	}
	logLevel.Store(int32(level))
}

func resetLevel(ver int64) {
	if GetLevel() < InfoLvl && resetVer.Load() == ver {
		SetLevel(initialLogLevel, 0)
		return
	}
}

func GetLevel() zapcore.Level {
	return zapcore.Level(int8(logLevel.Load()))
}

func initDefaultLogger(config *Config) {
	if config.LogFileName == "" {
		config.LogFileName = DefaultLogFileName
	}
	for _, name := range nameMap {
		if config.LogFileName == name {
			config.LogFileName = DefaultLogFileName
		}
	}

	var opts []option
	printToStd := config.PrintToStd
	if (printToStd == PrintToStd_USERLOG || printToStd == PrintToStd_ALL) && !env.IsLive() {
		printToStdOut(config)
		return
	}

	splitLevel, ok := checkLevel(config.SplitLevel)

	// 获取分割log的option
	if ok {
		// 例如config.Level为InfoLvl，splitLevel小于config.Level时，也没有意义，小于config.Level的日志不会被记录
		if splitLevel < config.Level {
			splitLevel = config.Level
		}
		opts = getSplitOpt(config, splitLevel)
	} else {
		opts = getDefaultOpt(config)
	}

	logger = newLogger(opts...)
	zap.ReplaceGlobals(logger)
}

func printToStdOut(config *Config) {
	opt := option{
		Stdout: true,
		Lef: func(level LogLevel) bool {
			return level >= GetLevel()
		},
	}
	logger = newLogger(opt)
}

func newLogger(opts ...option) *zap.Logger {
	var cores []zapcore.Core
	enCoderConfig := zaplib.NewProductionEncoderConfig()
	enCoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(customTimeLayout)
	enCoderConfig.EncodeDuration = zapcore.MillisDurationEncoder
	enCoderConfig.ConsoleSeparator = "|"
	encoder := zaplib.NewConsoleEncoder(enCoderConfig)

	for _, opt := range opts {
		core := newCore(encoder, opt)
		cores = append(cores, core)
	}

	logger := zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddStacktrace(zap.PanicLevel))
	logger = logger.With(zap.String(zaplib.TraceKey, "-"))

	return logger

}

func newCore(encoder zapcore.Encoder, opt option) zapcore.Core {
	lv := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return opt.Lef(lvl)
	})

	var syncer io.Writer
	if opt.Stdout {
		syncer = os.Stdout
	} else {
		syncer = &lumberjack.Logger{
			Filename:   opt.FileName,
			MaxSize:    opt.Ropt.MaxSize,
			MaxAge:     opt.Ropt.MaxAge,
			MaxBackups: opt.Ropt.MaxBackups,
			LocalTime:  opt.LocalTime,
			Compress:   opt.Ropt.Compress,
		}
	}
	w := zapcore.AddSync(syncer)
	core := zapcore.NewCore(encoder, w, lv)
	return core
}

func checkLevel(splitLevel SplitLevel) (LogLevel, bool) {
	if splitLevel == "" || splitLevel == SplitNone {
		return 0, false
	}
	if level, ok := levelMap[splitLevel]; ok {
		return level, true
	}
	return 0, false
}

func initSystemLogger(config *Config) {
	var opts []option
	printToStd := config.PrintToStd
	// 如果在非线上环境设置了printToStd
	if (printToStd == PrintToStd_SYSLOG || printToStd == PrintToStd_ALL) && !env.IsLive() {
		opts = append(opts, option{
			Stdout: true,
			Lef: func(level zapcore.Level) bool {
				return level >= GetLevel()
			},
		})
	} else {
		// 将错误级别日志输出到 SysErrorLogFileName
		opts = append(opts, getOption(config, SysErrorLogFileName, func(level zapcore.Level) bool {
			return level >= ErrorLvl
		}))
		// 将大于等于config.Level级别的日志输出至 SysLogFileName
		opts = append(opts, getOption(config, SysLogFileName, func(level zapcore.Level) bool {
			return level >= GetLevel()
		}))
	}

	sysLogger = newLogger(opts...)
	// 替换gRPC库的日志记录器
	grpczap.ReplaceGrpcLoggerV2(sysLogger)
}

func initTracingLogger(config *Config) {
	printToStd := config.PrintToStd
	if config.TracingLogFileName == "" {
		config.TracingLogFileName = DefaultTracingFileName
	}
	for _, fileName := range nameMap {
		if config.TracingLogFileName == fileName {
			config.TracingLogFileName = DefaultTracingFileName
		}
	}
	var opts []option
	if (printToStd == PrintToStd_TRACING || printToStd == PrintToStd_ALL) && !env.IsLive() {
		// 将大于等于config.Level级别的日志输出至 Stdout
		opts = append(opts, option{
			Stdout: true,
			Lef: func(level zapcore.Level) bool {
				return level >= GetLevel()
			},
		})
	} else {
		// 将大于等于config.Level级别的日志输出至 config.TracingLogFileName
		opts = append(opts, getOption(config, config.TracingLogFileName, func(level zapcore.Level) bool {
			return level >= GetLevel()
		}))
	}

	tracingLogger = newLogger(opts...)
}

type rotateOptions struct {
	MaxSize    int
	MaxAge     int
	MaxBackups int
	Compress   bool
}

type option struct {
	LocalTime bool
	Stdout    bool
	FileName  string
	Ropt      rotateOptions
	Lef       zap.LevelEnablerFunc
}

func getSplitOpt(config *Config, splitLevel LogLevel) []option {
	var opts []option

	// server.log
	if splitLevel != DebugLvl {
		// 匿名函数 判断日志级别是否在当前级别（通过GetLevel()获取）和splitLevel之间
		opts = append(opts, getOption(config, config.LogFileName, func(level zapcore.Level) bool {
			return level >= GetLevel() && level <= splitLevel
		}))
	}

	// 分割日志到不同的文件
	for level, fileName := range nameMap {
		if level > splitLevel && level != ErrorLvl {
			l := level
			// 匿名函数 判断日志级别是否在当前级别（通过GetLevel()获取）且等于当前遍历的级别。
			opts = append(opts, getOption(config, fileName, func(lvl zapcore.Level) bool {
				return lvl >= GetLevel() && lvl == l
			}))
		}
	}

	// error.log 包含了所有日志级别大于Error的日志
	opts = append(opts, getOption(config, nameMap[ErrorLvl], func(level zapcore.Level) bool {
		return level >= GetLevel() && level >= ErrorLvl
	}))

	return opts
}

func getOption(config *Config, logFileName string, enablerFunc zap.LevelEnablerFunc) option {
	return option{
		FileName: env.GetFilePath(config.Path, logFileName),
		Ropt: rotateOptions{
			MaxSize:    100,
			MaxAge:     7,
			MaxBackups: 10,
			Compress:   config.Compress,
		},
		Lef: enablerFunc,
	}
}

func getDefaultOpt(config *Config) []option {
	var opts []option
	opts = append(opts, getOption(config, config.LogFileName, func(level zapcore.Level) bool {
		return level >= GetLevel()
	}))
	return opts
}

// SetLogFileName 设置按级别拆分的日志文件名。
// 如果newName有效且设置成功返回true，如果newName与其他日志文件重复且设置失败返回false。
// 应在记录器初始化之前调用。
// 例如SetLogFileName(log.DebugLvl，"my_debug")，那么所有的调试日志将写入my_debug.log。
func SetLogFileName(level LogLevel, newName string) bool {
	if !checkLogFileNameValid(level, newName) {
		return false
	}
	nameMap[level] = newName
	return true
}

func checkLogFileNameValid(level LogLevel, newName string) bool {
	if newName == "" || newName == SysLogFileName || newName == SysErrorLogFileName || newName == DefaultLogFileName || newName == DefaultTracingFileName {
		return false
	}
	for l, s := range nameMap {
		if l != level && newName == s {
			return false
		}
	}
	return true
}
