package setting

import (
	"log"
	"strings"
	"time"

	"github.com/go-ini/ini"
)

type App struct {
	LogSavePath  string
	LogSaveName  string
	LogLevel     string
	MaxMsgSize   int64
	MemQueueSize int64
	ArgumentType string
}

type Server struct {
	BindAddress  string
	ServerIP     string
	ServerPort   int
	Auth         bool
	AuthUsername string
	AuthPassword string
	AuthModel    string
}
type Etcd struct {
	ConnStr string
	Enable  bool
	Timeout int
	Dispatch bool
	DispatchTopic string //调度通道
	Broadcast string  //广播通道
}
type Http struct {
	Port         int
	Enable       bool
	RunMode      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}
type Authorization struct {
	Enable bool
}
type Running struct {
	IsAuthorization   bool
	AuthorizationCode string
	IgnoreTopic       []string
}

var (
	cfg                  *ini.File
	AppSetting           = &App{}
	ServerSetting        = &Server{}
	EtcdSetting          = &Etcd{}
	HttpSetting          = &Http{}
	AuthorizationSetting = &Authorization{}
	RunningSetting       = &Running{
		IsAuthorization: false,
	}
)

func InitSetting(path string) (err error) {
	if cfg, err = ini.Load(path); err != nil {
		return
	}

	mapTo("app", AppSetting)
	mapTo("server", ServerSetting)
	mapTo("etcd", EtcdSetting)
	mapTo("http", HttpSetting)
	mapTo("authorization", AuthorizationSetting)
	IgnoreTopic := cfg.Section("app").Key("IgnoreTopic").String()
	RunningSetting.IgnoreTopic = strings.Split(IgnoreTopic, ",")
	return
}
func mapTo(section string, v interface{}) {
	err := cfg.Section(section).MapTo(v)
	if err != nil {
		log.Fatalf("Cfg.MapTo RedisSetting err: %v", err)
	}
}
