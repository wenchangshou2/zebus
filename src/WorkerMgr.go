package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/wenchangshou2/zebus/src/pkg/e"
	"github.com/wenchangshou2/zebus/src/pkg/logging"
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

type WorkerMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
	hub    *Hub
}

var (
	G_workerMgr *WorkerMgr
)
// 初始经状态同步
func InitWorkerMgr(hub *Hub) (err error) {
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
	G_workerMgr = &WorkerMgr{
		client: client,
		kv:     kv,
		lease:  lease,
		hub:hub,
	}

	logging.G_Logger.Info("info workermgr success")
	return
}
func (WorkerMgr *WorkerMgr) ListWorkers() (workerArr []e.WorkerInfo, err error) {
	var (
		getResp  *clientv3.GetResponse
		kv       *mvccpb.KeyValue
		workerIp string
	)
	workerArr = make([]e.WorkerInfo, 0)
	if getResp, err = WorkerMgr.kv.Get(context.TODO(), e.JOB_WORKER_DIR, clientv3.WithPrefix()); err != nil {
		return
	}
	for _, kv = range getResp.Kvs {
		if utils2.IsDaemon(string(kv.Key)) {
			workerIp = utils2.ExtractWorkerIP(string(kv.Key))
			serverInfo := e.WorkerInfo{
				Ip:     workerIp,
				Server: make([]string, 0),
			}
			workerArr = append(workerArr, serverInfo)
		} else {
			workerIp, serverName := utils2.ExtractServerName(string(kv.Key))
			for idx, server := range workerArr {
				if strings.Compare(server.Ip, workerIp) == 0 {
					workerArr[idx].Server = append(server.Server, serverName)
				}
			}
		}
	}
	return
}
//Daemon上线时调用，表示展期机器的上线
func (WorkerMgr *WorkerMgr) PutServerInfo(serverName string,serverType string) (err error) {
	var (
		topic string
	)
	logging.G_Logger.Info("new daemon client up,up topic:"+e.JOB_WORKER_DIR+serverName)
	if len(serverType)>0{
		topic=e.JOB_SERVER_DIR+serverName+"/"+serverType
	}else{
		topic=e.JOB_SERVER_DIR+serverName
	}
	fmt.Println("topic",topic)
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	_, err = WorkerMgr.client.Put(ctx, topic, serverType)
	if err != nil {
		logging.G_Logger.Warn("put  Host info fail:"+err.Error())
		return
	}
	return
}

//获取所有的客户端
func (WorkerMgr *WorkerMgr) GetAllClient() (clients []string, err error) {
	var (
		resp *clientv3.GetResponse
	)
	clients = make([]string, 0)
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	resp, err = WorkerMgr.client.Get(ctx, e.JOB_SERVER_DIR, clientv3.WithPrefix())
	if err != nil {
		logging.G_Logger.Warn("get cleitns error:" + err.Error())
		return clients, err
	}
	for _, ev := range resp.Kvs {
		key:=strings.ReplaceAll(string(ev.Key),"//","/")
		tmp := strings.Split(key, "/")
		clients = append(clients, tmp[2])
	}
	return
}
