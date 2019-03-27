package main

import (
	"fmt"
	"github.com/wenchangshou2/zebus/src/pkg/logging"
	"github.com/wenchangshou2/zebus/src/pkg/setting"
	"time"
)

func main(){
	var (
		err error
	)
	if err=setting.InitSetting("conf/app.ini");err!=nil{
		fmt.Println("读取配置文件失败")
		return
	}
	if err=logging.InitLogging(setting.AppSetting.LogSavePath,setting.AppSetting.LogLevel);err!=nil{
		fmt.Println("创建日志失败")
		return
	}

	serverAddr:=fmt.Sprintf("%s:%d",setting.ServerSetting.ServerIp,setting.ServerSetting.ServerPort)
	if err=InitSchedume(serverAddr);err!=nil{
		logging.G_Logger.Error("创建调度失败")
		return
	}
	
	fmt.Println("1111")
	if err=InituPnpServer("0.0.0.0",8888);err!=nil{
		logging.G_Logger.Error("创建 pnp失败")
		return
	}
	for{
		time.Sleep(1*time.Second)
	}

}