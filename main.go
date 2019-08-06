package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/wenchangshou2/zebus/pkg/http_api"
	"log"
	"net"
	"os"

	"github.com/wenchangshou2/zebus/pkg/certification"

	//_ "net/http/pprof"

	"github.com/kardianos/service"
	"github.com/wenchangshou2/zebus/pkg/logging"
	"github.com/wenchangshou2/zebus/pkg/setting"
	"github.com/wenchangshou2/zebus/pkg/utils"
)

type Service struct {
}

func (*Service) Start(_ service.Service) error {
	var (
		err               error
		AuthorizationDone chan bool
	)
	AuthorizationDone = make(chan bool)
	hub := newHub()
	confPath, _ := utils.GetFullPath("conf/app.ini")
	if err = setting.InitSetting(confPath); err != nil {
		fmt.Println("读取配置文件失败")
		return errors.New("读取配置文件失败")
	}
	logPath, _ := utils.GetFullPath(setting.AppSetting.LogSavePath)
	if err = logging.InitLogging(logPath, setting.AppSetting.LogLevel); err != nil {
		return errors.New("创建日志失败")
	}
	if err = certification.InitCertification(); err != nil { //初始化认证
		return errors.New("初始化授权失败")
	}

	httpServer:=newHTTPServer(hub,false,false)
	httpListener, err := net.Listen("tcp", "0.0.0.0:9191")
	go http_api.Serve(httpListener, httpServer, "HTTP", *logging.G_Logger)
	if setting.AuthorizationSetting.Enable {
		InitAuthorization(AuthorizationDone)
	}
	serverAddr := fmt.Sprintf("%s:%d", setting.ServerSetting.ServerIp, setting.ServerSetting.ServerPort)
	if err = InitSchedume(serverAddr, hub); err != nil {
		logging.G_Logger.Error("创建调度失败")
		return fmt.Errorf("创建调度失败")
	}

	if err = InituPnpServer("0.0.0.0", 8888); err != nil {
		return fmt.Errorf("创建pnp失败")
	}
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
