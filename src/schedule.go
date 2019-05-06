package main

import (
	"github.com/wenchangshou2/zebus/_vendor-20190426135752/github.com/pkg/errors"
	"github.com/wenchangshou2/zebus/src/pkg/setting"
	"net/http"
)

type ServerList struct {
}

func InitSchedume(addr string) (err error) {
	hub := newHub()
	go hub.run()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	if setting.EtcdSetting.Enable{
		if err=InitWorkerMgr(hub);err!=nil{
			return errors.New("创建etcd workerear失败")
		}
		if err=InitScheduleMgr(hub);err!=nil{
			return errors.New("创建etcd 同步服务失败")
		}
	}
	go http.ListenAndServe(addr, nil)
	return
}
