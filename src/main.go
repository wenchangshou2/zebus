package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	_ "net/http/pprof"

	"github.com/kardianos/service"
	"github.com/pkg/errors"
	"github.com/wenchangshou2/zebus/src/pkg/logging"
	"github.com/wenchangshou2/zebus/src/pkg/setting"
	"github.com/wenchangshou2/zebus/src/pkg/utils"
)

type Service struct {
}

func (*Service) Start(_ service.Service) error {
	var (
		err error
	)
	// time.Sleep(60 * time.Second)

	confPath, _ := utils.GetFullPath("conf/app.ini")
	if err = setting.InitSetting(confPath); err != nil {
		fmt.Println("读取配置文件失败")
		return errors.New("读取配置文件失败")
	}
	logPath, _ := utils.GetFullPath(setting.AppSetting.LogSavePath)
	fmt.Println("log path", logPath)
	if err = logging.InitLogging(logPath, setting.AppSetting.LogLevel); err != nil {
		fmt.Println("创建日志失败")
		return errors.New("创建日志失败")
	}

	serverAddr := fmt.Sprintf("%s:%d", setting.ServerSetting.ServerIp, setting.ServerSetting.ServerPort)
	if err = InitSchedume(serverAddr); err != nil {
		return errors.New("创建调试失败")
	}


	if err = InituPnpServer("0.0.0.0", 8888); err != nil {
		fmt.Println("创建pnp失败")
		return errors.New("创建pnp失败")
	}
	//fmt.Println("server", serverAddr)
	return nil
}
func (*Service) Stop(_ service.Service) error {
	return nil
}

var serviceFlag = flag.String("service", "", "Control the service")

func main() {
	var (
		err error
		s   service.Service
	)
	flag.Parse()

	svcConfig := &service.Config{
		Name:        "zoolon-zebus",
		DisplayName: "zoolon-zebus",
		Description: "zoolon 消息服务",
	}

	svc := &Service{}
	s, err = service.New(svc, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			s.Install()
			fmt.Println("服务安装成功")
			return

		case "remove":
			s.Uninstall()
			fmt.Println("服务卸载成功")
			return
		}
	}
	err = s.Run()
	if err != nil {
		fmt.Errorf("启动服务失败:%v", err)
	}

}
