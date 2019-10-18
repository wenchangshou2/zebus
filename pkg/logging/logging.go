package logging

import (
	"fmt"
	"net/url"
	"os"
	"time"
	"gopkg.in/natefinch/lumberjack.v2"
	"github.com/wenchangshou2/zebus/pkg/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logging struct {
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

	hook := lumberjack.Logger{
		Filename:   "./logs/zebus.log", // 日志文件路径
		MaxSize:    128,                      // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: 30,                       // 日志文件最多保存多少个备份
		MaxAge:     7,                        // 文件最多保存多少天
		Compress:   true,                     // 是否压缩
	}
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
	//fileSavePath := "winfile:///" + path.Join(logPath, formatted)
	zap.RegisterSink("winfile", newWinFileSink)

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),                                           // 编码器配置
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook)), // 打印到控制台和文件
		atom,                                                                     // 日志级别
	)
	caller := zap.AddCaller()
	development := zap.Development()
	filed := zap.Fields(zap.String("serviceName", "zebus"))
	// 构造日志
	//logger := zap.New(core, caller, development, filed)
	G_Logger=zap.New(core,caller,development,filed)
	if err != nil {
		fmt.Println("打开日志文件失败", err)
	}
	G_Logger.Info("log 初始化成功")

	return
}
