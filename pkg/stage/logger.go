package stage

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/sirupsen/logrus"
)

type PluginLogger struct {
	name   string
	logger logrus.FieldLogger
}

func NewPluginLogger() hclog.Logger {
	return &PluginLogger{
		logger: &logrus.Logger{
			Out:       os.Stderr,
			Level:     logrus.GetLevel(),
			Formatter: new(logrus.TextFormatter),
		},
	}
}

func (*PluginLogger) Trace(_ string, _ ...interface{}) {
	return
}

func (logger *PluginLogger) IsTrace() bool {
	return false
}

func (logger *PluginLogger) Debug(msg string, args ...interface{}) {
	// All plugin output logs to debug
	done := logger.upgradePluginOutput(msg, args)
	if !done {
		logger.logger.WithFields(argsToFields(args)).Debug(msg)
	}
}

func (logger *PluginLogger) IsDebug() bool {
	return logger.logger.WithFields(logrus.Fields{}).Level >= logrus.DebugLevel
}

func (logger *PluginLogger) Info(msg string, args ...interface{}) {
	logger.logger.WithFields(argsToFields(args)).Info(msg)
}

func (logger *PluginLogger) IsInfo() bool {
	return logger.logger.WithFields(logrus.Fields{}).Level >= logrus.InfoLevel
}

func (logger *PluginLogger) Warn(msg string, args ...interface{}) {
	logger.logger.WithFields(argsToFields(args)).Warn(msg)
}

func (logger *PluginLogger) IsWarn() bool {
	return logger.logger.WithFields(logrus.Fields{}).Level >= logrus.WarnLevel
}

func (logger *PluginLogger) Error(msg string, args ...interface{}) {
	logger.logger.WithFields(argsToFields(args)).Error(msg)
}

func (logger *PluginLogger) IsError() bool {
	return logger.logger.WithFields(logrus.Fields{}).Level >= logrus.ErrorLevel
}

func (logger *PluginLogger) SetLevel(_ hclog.Level) {}

func (logger *PluginLogger) With(args ...interface{}) hclog.Logger {
	return &PluginLogger{logger: logger.logger.WithFields(argsToFields(args))}
}

func (logger *PluginLogger) Named(name string) hclog.Logger {
	if len(logger.name) > 0 {
		return logger.ResetNamed(fmt.Sprintf("%s.%s", logger.name, name))
	}
	return logger.ResetNamed(name)
}

func (logger *PluginLogger) ResetNamed(name string) hclog.Logger {
	return &PluginLogger{name: name, logger: logger.logger.WithFields(logrus.Fields{"name": name})}
}

func (logger *PluginLogger) StandardLogger(_ *hclog.StandardLoggerOptions) *log.Logger {
	return log.New(logger.logger.WithFields(logrus.Fields{}).WriterLevel(logrus.InfoLevel), "", 0)
}

func (logger *PluginLogger) StandardWriter(_ *hclog.StandardLoggerOptions) io.Writer {
	if l, ok := logger.logger.(*logrus.Logger); ok {
		return l.Out
	}
	return os.Stderr
}

func (logger *PluginLogger) upgradePluginOutput(originalMsg string, args []interface{}) bool {
	var output map[string]string
	if err := json.Unmarshal([]byte(originalMsg), &output); err != nil {
		return false
	}

	msg, msgOk := output["msg"]
	level, levelOk := output["level"]
	if !msgOk || !levelOk {
		return false
	}

	fields := argsToFields(args)
	for k, v := range output {
		if k != "msg" && k != "level" {
			fields[k] = v
		}
	}

	switch level {
	case "debug":
		logger.logger.WithFields(fields).Debug(msg)
	case "info":
		logger.logger.WithFields(fields).Info(msg)
	case "warn":
		logger.logger.WithFields(fields).Warn(msg)
	case "error":
		logger.logger.WithFields(fields).Error(msg)
	}
	return true
}

func argsToFields(args []interface{}) logrus.Fields {
	if len(args)%2 != 0 {
		args = append(args, "")
	}
	fields := make(logrus.Fields, len(args)/2)
	for i := 0; i < len(args); i = i + 2 {
		if key, ok := args[i].(string); ok {
			fields[key] = args[i+1]
		}
	}
	return fields
}
