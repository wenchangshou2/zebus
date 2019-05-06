package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/wenchangshou2/zebus/src/pkg/e"
	"github.com/wenchangshou2/zebus/src/pkg/setting"
	utils2 "github.com/wenchangshou2/zebus/src/pkg/utils"
	"go.etcd.io/etcd/clientv3"
	"strings"
	"time"
)

var (
	G_etcd_client *clientv3.Client
	config        clientv3.Config
)
type WorkerMgr struct{
	client *clientv3.Client
	kv clientv3.KV
	lease clientv3.Lease
}
var (
	G_workerMgr *WorkerMgr
)


func InitWorkerMgr() (err error){
	var (
		config clientv3.Config
		client *clientv3.Client
		kv clientv3.KV
		lease clientv3.Lease
	)
	config =clientv3.Config{Endpoints:[]string{setting.EtcdSetting.ConnStr},DialTimeout:5*time.Second}
	if client,err=clientv3.New(config);err!=nil{
		return
	}
	kv=clientv3.NewKV(client)
	lease=clientv3.NewLease(client)
	G_workerMgr=&WorkerMgr{
		client:client,
		kv:kv,
		lease:lease,
	}
	fmt.Println("init workermgr success")
	return
}
func (WorkerMgr *WorkerMgr)ListWorkers()(workerArr []e.WorkerInfo,err error){
	fmt.Println("list worker")
	var (
		getResp *clientv3.GetResponse
		kv *mvccpb.KeyValue
		workerIp string
	)
	workerArr=make([]e.WorkerInfo,0)
	if getResp,err=WorkerMgr.kv.Get(context.TODO(),e.JOB_WORKER_DIR,clientv3.WithPrefix());err!=nil{
		return
	}
	for _, kv = range getResp.Kvs {
		fmt.Println("kv",string(kv.Key))
		fmt.Println("is Daemon",string(kv.Key))

		if utils2.IsDaemon(string(kv.Key)){
			workerIp=utils2.ExtractWorkerIP(string(kv.Key))
			serverInfo:=e.WorkerInfo{
				Ip:workerIp,
				Server:make([]string,0),
			}
			fmt.Println("111",serverInfo,workerIp)
			workerArr=append(workerArr,serverInfo)
		}else{
			fmt.Println("is server")
			workerIp,serverName:=utils2.ExtractServerName(string(kv.Key))
			fmt.Println("info ",workerIp,serverName)
			for idx,server:=range workerArr{
				if strings.Compare(server.Ip,workerIp)==0{
					fmt.Println("yes")
					workerArr[idx].Server=append(server.Server,serverName)
				}
			}
		}
	}
	fmt.Println("server info",workerArr)
	return
}