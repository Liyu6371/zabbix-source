package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"zabbix-source/config"
	"zabbix-source/utils"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

var (
	logLevelMap = map[string]logrus.Level{
		"error": logrus.ErrorLevel,
		"warn":  logrus.WarnLevel,
		"info":  logrus.InfoLevel,
		"debug": logrus.DebugLevel,
	}

	maxSize    = 10 // 每个日志文件最大10MB
	maxBackups = 3  // 保留最近的3个日志文件
	maxAge     = 7  // 保留最近7天的日志

	defaultLogger *logrus.Logger
	defaultLevel  = logrus.DebugLevel
	defaultLogDir = "/var/log/gse/"
)

func Init(c config.LoggerConfig) {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logger.SetLevel(defaultLevel)
	if c.Level != "" {
		if v, ok := logLevelMap[c.Level]; ok {
			logger.SetLevel(v)
		}
	}

	var logWritePath string
	execName, err := utils.GetExecutableName()
	if err != nil {
		fmt.Println("unable to get Executable Name:", err)
		os.Exit(1)
	}
	if c.OutputPath != "" {
		logWritePath = filepath.Join(c.OutputPath, execName+".log")
	} else {
		logWritePath = filepath.Join(defaultLogDir, execName+".log")
	}
	logDir := filepath.Dir(logWritePath)
	if !utils.IsDirWritable(logDir) {
		fmt.Println("log directory is not writable:", logDir)
		os.Exit(1)
	}
	logger.SetOutput(&lumberjack.Logger{
		Filename:   logWritePath,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   true,
	})
	defaultLogger = logger
}

func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

func Debugln(args ...interface{}) {
	defaultLogger.Debugln(args...)
}

func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

func Infoln(args ...interface{}) {
	defaultLogger.Infoln(args...)
}

func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

func Warn(args ...interface{}) {
	defaultLogger.Warn(args...)
}

func Warnln(args ...interface{}) {
	defaultLogger.Warnln(args...)
}

func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

func Errorln(args ...interface{}) {
	defaultLogger.Errorln(args...)
}

func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}
