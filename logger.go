package logger

import (
	"strings"
	"time"
)

const (
	OutputConsole = "console"
	OutputFile    = "file"
)

type RollingFileConfig struct {
	MaxSize    int  `json:"maxSize"`    //每个日志文件保存的最大尺寸 单位：M
	MaxBackups int  `json:"maxBackups"` //日志文件最多保存多少个备份
	MaxAge     int  `json:"maxAge"`     //文件最多保存多少天
	Compress   bool `json:"compress"`   //是否压缩
}

type OutputConfig struct {
	Type     string
	FileName string
	*RollingFileConfig
}

func NewOutputConsole() *OutputConfig {
	return &OutputConfig{
		Type: OutputConsole,
	}
}

func NewOutputFile(path string, fileName string, r *RollingFileConfig) *OutputConfig {
	filePath := path
	if !strings.HasSuffix(path, "/") {
		filePath += "/"
	}
	filePath += fileName

	return &OutputConfig{
		Type:              OutputFile,
		FileName:          filePath,
		RollingFileConfig: r,
	}
}

type Config struct {
	Outputs []*OutputConfig
	Fields  []*Field
}

type Option func(*Config)

func WithConsole() Option {
	return func(opts *Config) {
		opts.Outputs = append(opts.Outputs, NewOutputConsole())
	}
}
func WithFile(path string, fileName string, r *RollingFileConfig) Option {
	return func(opts *Config) {
		opts.Outputs = append(opts.Outputs, NewOutputFile(path, fileName, r))
	}
}
func WithOutputs(val ...*OutputConfig) Option {
	return func(opts *Config) {
		opts.Outputs = append(opts.Outputs, val...)
	}
}
func WithFields(val ...*Field) Option {
	return func(opts *Config) {
		opts.Fields = append(opts.Fields, val...)
	}
}
func WithField(key string, value interface{}) Option {
	return func(opts *Config) {
		opts.Fields = append(opts.Fields, NewField(key, value))
	}
}

type Field struct {
	Key   string
	Value interface{}
}

func NewField(key string, value interface{}) *Field {
	return &Field{Key: key, Value: value}
}

type Logger interface {
	Level() string
	Field(key string, value interface{}) *Field
	Debug(msg string, fields ...*Field)
	Info(msg string, fields ...*Field)
	Warn(msg string, fields ...*Field)
	Error(msg string, fields ...*Field)
	Fatal(msg string, fields ...*Field)
}

func MapToFields(data map[string]interface{}) []*Field {
	var fields []*Field
	for key, value := range data {
		fields = append(fields, &Field{key, value})
	}
	return fields
}

func WithTime(data ...*Field) []*Field {
	var timeExists bool
	for _, field := range data {
		if field.Key == "time" && field.Value != nil {
			timeExists = true
		}
	}
	if !timeExists {
		data = append(data, NewField("time", time.Now()))
	}
	return data
}
