package logger

import (
	"github.com/go-tron/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"strings"
)

func NewZapWithConfig(c *config.Config, name string, level string, opts ...Option) Logger {
	name = strings.ToLower(name)
	opts = append(opts, WithFile(c.GetString("logging.path"), name+".log", &RollingFileConfig{
		MaxSize:    c.GetInt("logging.maxSize"),
		MaxBackups: c.GetInt("logging.maxBackups"),
		MaxAge:     c.GetInt("logging.maxAge"),
		Compress:   c.GetBool("logging.compress"),
	}))
	if c.GetBool("logging.console") {
		opts = append(opts, WithConsole())
	}

	opts = append(opts, WithFields(
		NewField("app_name", strings.ToLower(c.GetString("application.name"))),
		NewField("app_env", strings.ToLower(c.GetString("application.env"))),
	))
	if c.GetString("cluster.namespace") != "" {
		opts = append(opts, WithFields(
			NewField("namespace", c.GetString("cluster.namespace")),
			NewField("node_name", c.GetString("cluster.nodeName")),
			NewField("pod_name", c.GetString("cluster.podName")),
		))
	}
	return NewZap(name, level, opts...)
}

func NewZap(name string, level string, opts ...Option) Logger {
	c := &Config{}
	for _, apply := range opts {
		apply(c)
	}

	if len(c.Outputs) == 0 {
		c.Outputs = append(c.Outputs, NewOutputConsole())
	}
	core := Core(level, c.Outputs...)

	c.Fields = append(c.Fields, NewField("logger_name", name))
	zapFields := ZapFields(c.Fields...)

	zapOpts := []zap.Option{
		zap.Fields(zapFields...),
	}

	return &ZapLogger{
		level:  level,
		logger: zap.New(core, zapOpts...),
	}
}

func ZapEncoder() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		//TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,  //小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,     //ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder, //
		EncodeCaller:   zapcore.FullCallerEncoder,      //全路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}
}

func ZapLevel(level string) zap.AtomicLevel {
	var zapLevel zapcore.Level
	switch strings.ToLower(level) {
	case "debug":
		zapLevel = zap.DebugLevel
	case "info":
		zapLevel = zap.InfoLevel
	case "warn":
		zapLevel = zap.WarnLevel
	case "error":
		zapLevel = zap.ErrorLevel
	case "panic":
		zapLevel = zap.PanicLevel
	case "fatal":
		zapLevel = zap.FatalLevel
	default:
		zapLevel = zap.InfoLevel
	}
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zapLevel)
	return atomicLevel
}

func RollingFile(conf *OutputConfig) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   conf.FileName,   //日志文件路径
		MaxSize:    conf.MaxSize,    //每个日志文件保存的最大尺寸 单位：M
		MaxBackups: conf.MaxBackups, //日志文件最多保存多少个备份
		MaxAge:     conf.MaxAge,     //文件最多保存多少天
		Compress:   conf.Compress,   //是否压缩
	}
}

func Core(level string, outputs ...*OutputConfig) zapcore.Core {
	var writeSyncers []zapcore.WriteSyncer
	for _, output := range outputs {
		if output.Type == OutputConsole {
			writeSyncers = append(writeSyncers, zapcore.AddSync(os.Stdout))
		}
		if output.Type == OutputFile {
			writeSyncers = append(writeSyncers, zapcore.AddSync(RollingFile(output)))
		}
	}

	return zapcore.NewCore(
		zapcore.NewJSONEncoder(ZapEncoder()),
		zapcore.NewMultiWriteSyncer(writeSyncers...),
		ZapLevel(level),
	)
}

func MapToZapFields(data map[string]interface{}) []zap.Field {
	var fields []zap.Field
	for key, value := range data {
		if value == nil {
			continue
		}
		var zapField zap.Field
		switch val := value.(type) {
		case []byte:
			zapField = zap.ByteString(key, val)
		default:
			zapField = zap.Any(key, val)
		}
		fields = append(fields, zapField)
	}
	return fields
}

func ZapFields(data ...*Field) []zap.Field {
	var fields []zap.Field
	for _, field := range data {
		if field.Value == nil {
			continue
		}
		var zapField zap.Field
		switch val := field.Value.(type) {
		case []byte:
			zapField = zap.ByteString(field.Key, val)
		default:
			zapField = zap.Any(field.Key, val)
		}
		fields = append(fields, zapField)
	}
	return fields
}

func ZapFieldsWithTime(data ...*Field) []zap.Field {
	return ZapFields(WithTime(data...)...)
}

type ZapLogger struct {
	level  string
	logger *zap.Logger
}

func (l *ZapLogger) Level() string {
	return l.level
}
func (l *ZapLogger) Field(key string, value interface{}) *Field {
	return NewField(key, value)
}
func (l *ZapLogger) Debug(msg string, fields ...*Field) {
	l.logger.Debug(msg, ZapFieldsWithTime(fields...)...)
}
func (l *ZapLogger) Info(msg string, fields ...*Field) {
	l.logger.Info(msg, ZapFieldsWithTime(fields...)...)
}
func (l *ZapLogger) Warn(msg string, fields ...*Field) {
	l.logger.Warn(msg, ZapFieldsWithTime(fields...)...)
}
func (l *ZapLogger) Error(msg string, fields ...*Field) {
	l.logger.Error(msg, ZapFieldsWithTime(fields...)...)
}
func (l *ZapLogger) Fatal(msg string, fields ...*Field) {
	l.logger.Fatal(msg, ZapFieldsWithTime(fields...)...)
}
