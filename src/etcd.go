package main

import (
	"fmt"
	"time"

	"github.com/wenchangshou2/zebus/src/pkg/setting"
	"go.etcd.io/etcd/clientv3"
)

var (
	G_etcd_client *clientv3.Client
	config        clientv3.Config
)

func InitEtcd() (err error){
	fmt.Println("etcd config",setting.EtcdSetting)
	config = clientv3.Config{
		Endpoints:   []string{setting.EtcdSetting.ConnStr},
		DialTimeout: 5 * time.Second,
	}
	if G_etcd_client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}
	G_etcd_client = G_etcd_client
	return
}
