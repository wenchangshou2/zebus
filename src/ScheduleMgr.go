package main

import (
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/segmentio/objconv/json"
	"github.com/wenchangshou2/zebus/src/pkg/e"
	"github.com/wenchangshou2/zebus/src/pkg/setting"
	"github.com/wenchangshou2/zebus/src/pkg/utils"
	"time"
)

type ConfigMgr struct{
	client *clientv3.Client
	kv clientv3.KV
	lease clientv3.Lease
	ProcessData chan []byte
	hub *Hub
}
var (
	G_ScheduleMgr *ConfigMgr
)
func (scheduleMgr *ConfigMgr) Process(){
	for{
		select {
		case message, ok := <-scheduleMgr.ProcessData:
			fmt.Println("?ok",ok)
			if !ok{
				return
			}
			msg:=e.RequestCmd{}
			json.Unmarshal(message,&msg)
			if utils.IsDaemon(msg.ReceiverName){
				ip:=utils.ExtractWorkerIP(msg.ReceiverName)

				fmt.Println("ip",ip,msg.Action)
			}

			fmt.Println("message",string(message))

		}
	}
}
func InitScheduleMgr(hub *Hub)(err error){
	var (
		config clientv3.Config
		client *clientv3.Client
		kv clientv3.KV
		lease clientv3.Lease
	)
	config = clientv3.Config{Endpoints: []string{setting.EtcdSetting.ConnStr}, DialTimeout: 5 * time.Second}
	if client,err=clientv3.New(config);err!=nil{
		return
	}
	kv=clientv3.NewKV(client)
	lease=clientv3.NewLease(client)
	G_ScheduleMgr=&ConfigMgr{
		client:client,
		kv:kv,
		lease:lease,
		ProcessData:make(chan []byte),
		hub:hub,
	}
	go G_ScheduleMgr.Process()
	return
}