package logger

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"gin-app-start/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var globalLogger *zap.Logger

const (
	// DefaultLevel the default log level
	DefaultLevel = zapcore.InfoLevel

	// DefaultTimeLayout the default time layout;
	DefaultTimeLayout = time.RFC3339
)

// Option custom setup config
type Option func(*option)

type option struct {
	level          zapcore.Level     // 日志级别
	fields         map[string]string // 日志字段
	file           io.Writer         // 日志输出目标
	timeLayout     string            // 时间格式
	disableConsole bool              // 是否禁用控制台输出
}

// WithDebugLevel only greater than 'level' will output
func WithDebugLevel() Option {
	return func(opt *option) {
		opt.level = zapcore.DebugLevel
	}
}

// WithInfoLevel only greater than 'level' will output
func WithInfoLevel() Option {
	return func(opt *option) {
		opt.level = zapcore.InfoLevel
	}
}

// WithWarnLevel only greater than 'level' will output
func WithWarnLevel() Option {
	return func(opt *option) {
		opt.level = zapcore.WarnLevel
	}
}

// WithErrorLevel only greater than 'level' will output
func WithErrorLevel() Option {
	return func(opt *option) {
		opt.level = zapcore.ErrorLevel
	}
}

// WithField add some field(s) to log
func WithField(key, value string) Option {
	return func(opt *option) {
		opt.fields[key] = value
	}
}

// WithFileP write log to some file
func WithFileP(file string) Option {
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0766); err != nil {
		panic(err)
	}

	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0766)
	if err != nil {
		panic(err)
	}

	return func(opt *option) {
		opt.file = zapcore.Lock(f)
	}
}

// WithFileRotationP write log to some file with rotation
func WithFileRotationP(file string, maxSize, maxAge int) Option {
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0766); err != nil {
		panic(err)
	}

	return func(opt *option) {
		opt.file = &lumberjack.Logger{ // concurrent-safed
			Filename:   file,    // 文件路径
			MaxSize:    maxSize, // 单个文件最大尺寸，默认单位 M
			MaxBackups: 300,     // 最多保留 300 个备份
			MaxAge:     maxAge,  // 最大时间，默认单位 day
			LocalTime:  true,    // 使用本地时间
			Compress:   true,    // 是否压缩 disabled by default
		}
	}
}

// WithTimeLayout custom time format
func WithTimeLayout(timeLayout string) Option {
	return func(opt *option) {
		opt.timeLayout = timeLayout
	}
}

// WithDisableConsole WithEnableConsole write log to os.Stdout or os.Stderr
func WithDisableConsole() Option {
	return func(opt *option) {
		opt.disableConsole = true
	}
}

func Init(config *config.Config, opts ...Option) (*zap.Logger, error) {
	switch config.Log.Level {
	case "debug":
		opts = append(opts, WithDebugLevel())
	case "info":
		opts = append(opts, WithInfoLevel())
	case "warn":
		opts = append(opts, WithWarnLevel())
	case "error":
		opts = append(opts, WithErrorLevel())
	default:
		opts = append(opts, WithInfoLevel())
	}

	opt := &option{level: DefaultLevel, fields: make(map[string]string)}
	for _, f := range opts {
		f(opt)
	}

	timeLayout := DefaultTimeLayout
	if opt.timeLayout != "" {
		timeLayout = opt.timeLayout
	}

	// similar to zap.NewProductionEncoderConfig()
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",       // 时间戳字段名
		LevelKey:      "level",      // 日志级别字段名
		NameKey:       "logger",     // used by logger.Named(key); optional; useless
		CallerKey:     "caller",     // 调用者字段名
		MessageKey:    "msg",        // 日志消息字段名
		StacktraceKey: "stacktrace", // use by zap.AddStacktrace; optional; useless
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format(timeLayout))
		},
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 全路径编码器
	}

	// 创建json格式的日志编码器
	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

	// lowPriority usd by info\debug\warn
	// 低优先级过滤器 (lowPriority)
	// 过滤范围：info、debug、warn 级别
	// 保留的日志级别：级别 >= 配置级别 且 < 错误级别
	// 示例：如果配置为 info 级别，则 debug 级别日志会被过滤掉
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= opt.level && lvl < zapcore.ErrorLevel
	})

	// highPriority usd by error\panic\fatal
	// 高优先级过滤器 (highPriority)
	// 过滤范围：error、panic、fatal 级别
	// 保留的日志级别：级别 >= 配置级别 且 >= 错误级别
	// 特点：错误级别日志总是会被记录，不受配置级别影响
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= opt.level && lvl >= zapcore.ErrorLevel
	})

	stdout := zapcore.Lock(os.Stdout) // lock for concurrent safe
	stderr := zapcore.Lock(os.Stderr) // lock for concurrent safe

	// 创建一个空的日志核心，准备接收多个输出目标
	core := zapcore.NewTee()

	// 控制台日志
	if !opt.disableConsole {
		// 日志多路输出
		core = zapcore.NewTee(
			// 普通日志输出到stdout
			zapcore.NewCore(jsonEncoder,
				zapcore.NewMultiWriteSyncer(stdout),
				lowPriority,
			),

			// 错误日志输出到stderr
			zapcore.NewCore(jsonEncoder,
				zapcore.NewMultiWriteSyncer(stderr),
				highPriority,
			),
		)
	}

	// 文件日志
	if opt.file != nil {
		core = zapcore.NewTee(core,
			zapcore.NewCore(jsonEncoder,
				zapcore.AddSync(opt.file), // 将文件写入器转换为zap兼容的同步器(普通文件或轮转文件写入器)

				// 保留的日志级别：级别 >= 配置级别
				zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
					return lvl >= opt.level
				}),
			),
		)
	}

	// 创建日志记录器
	logger := zap.New(core,
		zap.AddCaller(),         // 自动记录每条日志的调用者信息（内容包括文件名和行号）
		zap.ErrorOutput(stderr), // 指定Logger内部错误的输出目标
	)

	// 为所有日志记录自动添加预定义字段
	for key, value := range opt.fields {
		logger = logger.WithOptions(zap.Fields(zapcore.Field{Key: key, Type: zapcore.StringType, String: value}))
	}

	globalLogger = logger
	return logger, nil
}

func GetLogger() *zap.Logger {
	if globalLogger == nil {
		logger, _ := zap.NewDevelopment()
		globalLogger = logger
	}
	return globalLogger
}

func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

func With(fields ...zap.Field) *zap.Logger {
	return GetLogger().With(fields...)
}

func WithContext(fields ...zap.Field) *zap.Logger {
	return GetLogger().With(fields...)
}

func Close() {
	if globalLogger != nil {
		_ = globalLogger.Sync()
	}
}

// func InitDefault() {
// 	if globalLogger == nil {
// 		logger, _ := zap.NewDevelopment()
// 		globalLogger = logger
// 	}
// }

// func init() {
// 	if globalLogger == nil {
// 		config := zap.NewDevelopmentConfig()
// 		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
// 		logger, err := config.Build(zap.AddCallerSkip(1))
// 		if err != nil {
// 			logger, _ = zap.NewDevelopment()
// 		}
// 		globalLogger = logger
// 	}
// }
