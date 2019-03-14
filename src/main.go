package main

import (
	"fmt"
	"github.com/wenchangshou2/zebus/src/pkg/logging"
	"github.com/wenchangshou2/zebus/src/pkg/setting"
)

func main(){
	var (
		err error
	)
	if err=setting.InitSetting("conf/app.ini");err!=nil{
		goto ERR
	}
	if err=logging.InitLogging(setting.AppSetting.LogSavePath,setting.AppSetting.LogLevel);err!=nil{
		fmt.Println("err",err)
		goto ERR
	}
	if err=InitSchedume(":9090");err!=nil{
		goto ERR
	}
	return
	ERR:
		fmt.Println("启动失败",err)
}