package setting

import (
	"github.com/go-ini/ini"
	"log"
)
type App struct{
	LogSavePath string
	LogSaveName string
	LogLevel string
}
type Server struct{
	ServerIp string
	ServerPort int
}
var (
	cfg *ini.File
	AppSetting=&App{}
	ServerSetting=&Server{}
)
func InitSetting(path string)(err error){
	if cfg,err=ini.Load(path);err!=nil{
		return
	}
	mapTo("app",AppSetting)
	mapTo("server",ServerSetting)
	return
}
func mapTo(section string, v interface{}){
	err := cfg.Section(section).MapTo(v)
	if err != nil {
		log.Fatalf("Cfg.MapTo RedisSetting err: %v", err)
	}
}