package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	_ "net/http/pprof"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/kardianos/service"
	"github.com/wenchangshou2/zebus/pkg/certification"
	"github.com/wenchangshou2/zebus/pkg/http_api"
	"github.com/wenchangshou2/zebus/pkg/logging"
	"github.com/wenchangshou2/zebus/pkg/setting"
	"github.com/wenchangshou2/zebus/pkg/utils"
)

type Service struct {
	opts         Options
	httpListener net.Listener
	httpServer   *httpServer
	hub          *ZEBUSD
}

func (s *Service) InitHttpServer() error {
	var (
		err error
	)
	if setting.HttpSetting.Proto == "https" {
		// 配置https
		tlsConfig, err := s.BuildTLSConfig()
		if err != nil {
			panic(fmt.Sprintf("Failed to build TLS config - %s", err))
		}
		if tlsConfig == nil {
			panic("cert or key error")
		}
		s.httpListener, err = tls.Listen("tcp", "0.0.0.0:9191", tlsConfig)
		if err != nil {
			panic(fmt.Sprintf("listen (%s) failed - %s", "0.0.0.0:9191", err.Error()))
		}
		s.httpServer, _ = newHTTPServer(s.hub, true, true)
	} else {
		// 配置http
		if s.httpListener, err = net.Listen("tcp", "0.0.0.0:9191"); err != nil {
			return err
		}
		if s.httpServer, err = newHTTPServer(s.hub, false, false); err != nil {
			return err
		}
	}
	go http_api.Serve(s.httpListener, s.httpServer, setting.HttpSetting.Proto, *logging.G_Logger)
	return nil
}
func (s *Service) Start(_ service.Service) error {
	var (
		err               error
		AuthorizationDone chan bool
	)
	err = utils.CheckLicense()
	if err != nil {
		return errors.New("当前授权失败:" + err.Error())
	}

	AuthorizationDone = make(chan bool)
	confPath, _ := utils.GetFullPath("conf/app.ini")
	if err = setting.InitSetting(confPath); err != nil {
		panic("读取配置文件失败")
	}
	if setting.AppSetting.ArgumentType == "cmd" {
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
	s.hub = hub
	if err := s.InitHttpServer(); err != nil {
		panic("创建http服务错误")
	}
	if setting.AuthorizationSetting.Enable {
		_ = InitAuthorization(AuthorizationDone)
	}
	serverAddr := fmt.Sprintf(setting.ServerSetting.BindAddress)
	if err = InitSchedule(serverAddr, hub, s); err != nil {
		logging.G_Logger.Error("创建调度失败")
		return errors.New("创建调试失败")
	}
	if err = InitJobMgr(); err != nil {
		logging.G_Logger.Error("创建jobMgr失败:" + err.Error())
		return errors.New("创建JobMgr失败")
	}
	if err = InitUPNPServer("0.0.0.0", 8888); err != nil {
		return errors.New("创建pnp失败")
	}
	return nil
}
func (s *Service) SetRunningArguments() {
	setting.ServerSetting.BindAddress = s.opts.ServerBindAddress
	setting.ServerSetting.ServerIP = s.opts.ServerAddress
	setting.EtcdSetting.Enable = s.opts.EtcdEnable
	setting.EtcdSetting.ConnStr = s.opts.EtcdServer
}

func (*Service) Stop(_ service.Service) error {
	return nil
}

// BuildTLSConfig 生成tls配置
func (s *Service) BuildTLSConfig() (*tls.Config, error) {
	var (
		tlsConfig *tls.Config
		tlsCert   string
		err       error
	)
	if tlsCert, err = utils.GetFullPath(setting.HttpSetting.Cert); err != nil {
		return nil, err
	}
	//caCert, _ := ioutil.ReadFile(setting.HttpSetting.Cert)
	//caCertPool := x509.NewCertPool()
	//caCertPool.AppendCertsFromPEM(caCert)
	tlsKey, err := utils.GetFullPath(setting.HttpSetting.Key)
	if err != nil {
		return nil, err
	}
	tlsClientAuthPolicy := tls.VerifyClientCertIfGiven
	cert, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
	if err != nil {
		return nil, err
	}
	switch setting.HttpSetting.Policy {
	case "require":
		tlsClientAuthPolicy = tls.RequireAnyClientCert
	case "require-verify":
		tlsClientAuthPolicy = tls.RequireAndVerifyClientCert
	default:
		tlsClientAuthPolicy = tls.NoClientCert
	}
	tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tlsClientAuthPolicy,
		//ClientAuth:   tls.RequireAndVerifyClientCert,
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS10,
		MaxVersion:         tls.VersionTLS12,
	}

	tlsConfig.BuildNameToCertificate()
	return tlsConfig, nil
}

func main() {
	var (
		err  error
		s    service.Service
		args []string
	)
	//opts:=NewOptions()
	//flagSet:=SyncFlagSet(opts)
	//flagSet.Parse(os.Args[1:])
	//fmt.Println("vv",flagSet.Lookup("etcd-server"))
	var opts Options
	args, err = flags.ParseArgs(&opts, os.Args[1:])
	fmt.Println("args", args)
	if err != nil {
		fmt.Printf("参数解析错误:%s\n", err.Error())
		panic("参数解析错误")
	}

	svcConfig := &service.Config{
		Name:        "zoolon-zebus",
		DisplayName: "zoolon-zebus",
		Description: "zoolon 消息服务",
	}

	svc := &Service{
		opts: opts,
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
			code, err := utils.GetSystemUUID()
			if err != nil {
				log.Fatalf("获取系统唯一码失败:%s", err.Error())
				return
			}
			fmt.Printf("获取系统唯一码成功:%s\n", code)
			os.Exit(1)
			return
		}

	}
	err = s.Run()
	if err != nil {
		log.Fatalf("启动服务失败:%v", err)
	}

}
