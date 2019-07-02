package setting

import (
	"log"
	"time"

	"github.com/go-ini/ini"
)

type App struct {
	LogSavePath string
	LogSaveName string
	LogLevel    string
}
type Server struct {
	ServerIp   string
	ServerPort int
}
type Etcd struct {
	ConnStr string
	Enable  bool
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

var (
	cfg                  *ini.File
	AppSetting           = &App{}
	ServerSetting        = &Server{}
	EtcdSetting          = &Etcd{}
	HttpSetting          = &Http{}
	AuthorizationSetting = &Authorization{}
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
	return
}
func mapTo(section string, v interface{}) {
	err := cfg.Section(section).MapTo(v)
	if err != nil {
		log.Fatalf("Cfg.MapTo RedisSetting err: %v", err)
	}
}
