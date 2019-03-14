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
var (
	cfg *ini.File
	AppSetting=&App{}
)
func InitSetting(path string)(err error){
	if cfg,err=ini.Load(path);err!=nil{
		return
	}
	mapTo("app",AppSetting)
	return
}
func mapTo(section string, v interface{}){
	err := cfg.Section(section).MapTo(v)
	if err != nil {
		log.Fatalf("Cfg.MapTo RedisSetting err: %v", err)
	}
}