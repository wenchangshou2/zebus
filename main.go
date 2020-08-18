package main

import (
	"errors"
	"flag"
	"fmt"

	//"github.com/mreiferson/go-options"
	"github.com/jessevdk/go-flags"
	"log"
	"net"
	"os"
	"github.com/wenchangshou2/zebus/pkg/http_api"
	_ "net/http/pprof"
	"github.com/wenchangshou2/zebus/pkg/certification"
	"github.com/kardianos/service"
	"github.com/wenchangshou2/zebus/pkg/logging"
	"github.com/wenchangshou2/zebus/pkg/setting"
	"github.com/wenchangshou2/zebus/pkg/utils"
)

type Service struct {
	flagSet *flag.FlagSet
	opts Options
}
func (s *Service) Start(_ service.Service) error {
	var (
		err               error
		AuthorizationDone chan bool
	)
	//err=utils.CheckLicense()
	//if err!=nil{
	//	return errors.New("当前授权失败:"+err.Error())
	//}

	AuthorizationDone = make(chan bool)
	confPath, _ := utils.GetFullPath("conf/app.ini")
	if err = setting.InitSetting(confPath); err != nil {
		panic("读取配置文件失败")
	}
	if setting.AppSetting.ArgumentType=="cmd"{
		s.SetRunningArguments()
	}
	logPath, _ := utils.GetFullPath(setting.AppSetting.LogSavePath)

	if err = logging.InitLogging(logPath, setting.AppSetting.LogLevel); err != nil {
		panic("创建日志文件失败")
	}
	if err = certification.InitCertification(); err != nil { //初始化认证
		return errors.New("初始化授权失败")
	}
	hub := newHub(logging.G_Logger)
	httpServer := newHTTPServer(hub, false, false)
	httpListener, err := net.Listen("tcp", "0.0.0.0:9191")
	if err != nil {
		logging.G_Logger.Error("初始化HTTP失败:" + err.Error())
		return errors.New("初始化http失败")
	}
	go http_api.Serve(httpListener, httpServer, "HTTP", *logging.G_Logger)
	if setting.AuthorizationSetting.Enable {
		_ = InitAuthorization(AuthorizationDone)
	}
	serverAddr := fmt.Sprintf("%s", setting.ServerSetting.BindAddress)
	if err = InitSchedume(serverAddr, hub); err != nil {
		logging.G_Logger.Error("创建调度失败")
		return errors.New("创建调试失败")
	}
	if err=InitJobMgr();err!=nil{
		logging.G_Logger.Error("创建jobMgr失败")
		return errors.New("创建JobMgr失败")
	}
	if err = InituPnpServer("0.0.0.0", 8888); err != nil {
		return errors.New("创建pnp失败")
	}
	return nil
}
func (s *Service)SetRunningArguments(){
	setting.ServerSetting.BindAddress=s.opts.ServerBindAddress
	setting.ServerSetting.ServerIP=s.opts.ServerAddress
	setting.EtcdSetting.Enable=s.opts.EtcdEnable
	setting.EtcdSetting.ConnStr=s.opts.EtcdServer

}

func (*Service) Stop(_ service.Service) error {
	return nil
}


func main() {
	var (
		err error
		s   service.Service
	)
	fmt.Println("args",os.Args[1:])
	//opts:=NewOptions()
	//flagSet:=SyncFlagSet(opts)
	//flagSet.Parse(os.Args[1:])
	//fmt.Println("vv",flagSet.Lookup("etcd-server"))
	var opts Options
	args, err := flags.ParseArgs(&opts, os.Args[1:])
	fmt.Println("args",args)
	if err!=nil{
		fmt.Errorf("参数解析错误:"+err.Error())
		panic("参数解析错误")
	}

	svcConfig := &service.Config{
		Name:        "zoolon-zebus",
		DisplayName: "zoolon-zebus",
		Description: "zoolon 消息服务",
	}

	svc := &Service{
		opts:opts,
	}
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
		case "get":
			code,err:=utils.GetSystemUUID()
			if err!=nil{
				fmt.Errorf("获取系统唯一码失败:%s",err.Error())
				return
			}
			fmt.Printf("获取系统唯一码成功:%s\n",code)
			os.Exit(1)
			return
		}

	}
	err = s.Run()
	if err != nil {
		fmt.Errorf("启动服务失败:%v", err)
	}

}
