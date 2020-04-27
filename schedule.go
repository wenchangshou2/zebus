package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/wenchangshou2/zebus/pkg/logging"
	"github.com/wenchangshou2/zebus/pkg/setting"
)

type ServerList struct {
}

func InitSchedume(addr string, hub *ZEBUSD) (err error) {
	var (
		retriesCount = 10
	)
	logging.G_Logger.Info("开始启动调度")
	go hub.run()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	if setting.EtcdSetting.Enable {
		logging.G_Logger.Info("正在配置etcd")

		for {
			if err = InitWorkerMgr(hub); err != nil {
				logging.G_Logger.Error("创建 etcd workerear 失败")
				goto exit
			}
			if err = InitScheduleMgr(hub); err != nil {
				logging.G_Logger.Error("init scheduleMgr 失败")
				goto exit
			}
			logging.G_Logger.Info("启动etcd WorkerMgr 成功")
			break
		exit:
			retriesCount--
			if retriesCount == 0 {
				return fmt.Errorf("启动 etcd 服务失败")
			}
			logging.G_Logger.Info(fmt.Sprintf("etcd 服务连接失败,当前还剩余连接次数%d", retriesCount))
			time.Sleep(10 * time.Second)
		}
	}
	logging.G_Logger.Info("启动 websocket server:" + addr)
	go http.ListenAndServe(addr, nil)
	return
}
