package logging

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/wenchangshou2/zebus/src/pkg/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logging struct{
	logging *zap.Logger
}
var (
	G_Logger *zap.Logger
)

func newWinFileSink(u *url.URL) (zap.Sink, error) {
	// Remove leading slash left by url.Parse()
	return os.OpenFile(u.Path[1:], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
}
func InitLogging(logPath string, level string) (err error) {
	var (
		atom zap.AtomicLevel
	)
	if !utils.IsExist(logPath) {
		os.MkdirAll(logPath, 0755)
	}
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder, // 全路径编码器
	}

	switch level {
	case "debug":
		atom = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		atom = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "error":
		atom = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		atom = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	t := time.Now()
	formatted := fmt.Sprintf("%d-%02d-%02d",
		t.Year(), t.Month(), t.Day())
	formatted = formatted + ".log"
	fileSavePath := "winfile:///" + path.Join(logPath, formatted)
	zap.RegisterSink("winfile", newWinFileSink)

	config := zap.Config{
		Level:         atom,          // 日志级别
		Development:   true,          // 开发模式，堆栈跟踪
		Encoding:      "json",        // 输出格式 console 或 json
		EncoderConfig: encoderConfig, // 编码器配置
		//InitialFields:    map[string]interface{}{"serviceName": "spikeProxy"}, // 初始化字段，如：添加一个服务器名称
		OutputPaths:      []string{"stdout", fileSavePath}, // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
		ErrorOutputPaths: []string{"stderr"},
	}

	G_Logger, err = config.Build()
	if err != nil {
		fmt.Println("打开日志文件失败", err)
	}
	G_Logger.Info("log 初始化成功")

	return
}
