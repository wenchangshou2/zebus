package main

import (
	"encoding/json"
	"fmt"
	"github.com/wenchangshou2/zebus/pkg/e"
	"github.com/wenchangshou2/zebus/pkg/setting"
	"github.com/wenchangshou2/zebus/pkg/utils"
	"go.etcd.io/etcd/clientv3"
	"time"
)

type ConfigMgr struct {
	client      *clientv3.Client
	kv          clientv3.KV
	lease       clientv3.Lease
	ProcessData chan []byte
	hub         *ZEBUSD
}

var (
	G_ScheduleMgr *ConfigMgr
)

func (scheduleMgr *ConfigMgr) Process() {
	for {
		select {
		case message, ok := <-scheduleMgr.ProcessData:
			if !ok {
				return
			}
			msg := e.RequestCmd{}
			json.Unmarshal(message, &msg)
			if utils.IsDaemon(msg.ReceiverName) {
				ip := utils.ExtractWorkerIP(msg.ReceiverName)
				fmt.Println("ip", ip, msg.Action)
			}
			fmt.Println("message", string(message))
		}
	}
}
func enitScheduleMgr(hub *ZEBUSD) (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
	)
	config = clientv3.Config{Endpoints: []string{setting.EtcdSetting.ConnStr}, DialTimeout: 5 * time.Second}
	if client, err = clientv3.New(config); err != nil {
		return
	}
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	G_ScheduleMgr = &ConfigMgr{
		client:      client,
		kv:          kv,
		lease:       lease,
		ProcessData: make(chan []byte),
		hub:         hub,
	}
	go G_ScheduleMgr.Process()
	return
}
