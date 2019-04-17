package setting

import (
	"log"

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
}

var (
	cfg           *ini.File
	AppSetting    = &App{}
	ServerSetting = &Server{}
	EtcdSetting   = &Etcd{}
)

func InitSetting(path string) (err error) {
	if cfg, err = ini.Load(path); err != nil {
		return
	}
	mapTo("app", AppSetting)
	mapTo("server", ServerSetting)
	mapTo("etcd", EtcdSetting)
	return
}
func mapTo(section string, v interface{}) {
	err := cfg.Section(section).MapTo(v)
	if err != nil {
		log.Fatalf("Cfg.MapTo RedisSetting err: %v", err)
	}
}
