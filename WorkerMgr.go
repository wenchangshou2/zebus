package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/wenchangshou2/zebus/pkg/e"
	"github.com/wenchangshou2/zebus/pkg/logging"
	"github.com/wenchangshou2/zebus/pkg/setting"
	utils2 "github.com/wenchangshou2/zebus/pkg/utils"
	"go.etcd.io/etcd/clientv3"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	G_etcd_client *clientv3.Client
	config        clientv3.Config
)

type WorkerMgr struct {
	client       *clientv3.Client
	kv           clientv3.KV
	lease        clientv3.Lease
	hub          *ZEBUSD
	clientsInfo  map[string]*e.ConfigInfo
	resourceInfo map[string][]string
}

var (
	G_workerMgr *WorkerMgr
)

// 初始经状态同步
func InitWorkerMgr(hub *ZEBUSD) (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
	)
	fmt.Println("setting.EtcdSetting.ConnStr",setting.EtcdSetting.ConnStr)
	config = clientv3.Config{Endpoints: []string{setting.EtcdSetting.ConnStr}, DialTimeout: 5 * time.Second}
	if client, err = clientv3.New(config); err != nil {
		return
	}
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	G_workerMgr = &WorkerMgr{
		client:       client,
		kv:           kv,
		lease:        lease,
		hub:          hub,
		clientsInfo:  make(map[string]*e.ConfigInfo),
		resourceInfo: make(map[string][]string),
	}
	go G_workerMgr.deployUpdateNotify()
	go G_workerMgr.ResourceUpdateNotify()
	logging.G_Logger.Info("info workermgr success")
	return
}
func (workerMgr *WorkerMgr) updateConfig(addr, item, val string) {
	var (
		client *e.ConfigInfo
		ok     bool
	)
	if client, ok = workerMgr.clientsInfo[addr]; ok {
		client = workerMgr.clientsInfo[addr]
	} else {
		client = &e.ConfigInfo{}
		workerMgr.clientsInfo[addr] = client
	}
	if strings.Compare(item, "volume") == 0 {
		client.Volume, _ = strconv.Atoi(val)
	}
	workerMgr.clientsInfo[addr]=client
}
func (workerMgr *WorkerMgr) updateResourceInfo(addr string, aid string) {
	var (
		resource []string
		ok       bool
	)
	if resource, ok = workerMgr.resourceInfo[addr]; ok {
		resource = workerMgr.resourceInfo[addr]
	} else {
		resource = make([]string, 0)
		workerMgr.resourceInfo[addr] = resource
	}
	for _, id := range resource {
		if id == aid {
			return
		}
	}
	workerMgr.resourceInfo[addr] = append(workerMgr.resourceInfo[addr], aid)
}
func (workerMgr *WorkerMgr) deleteResourceInfo(addr string,aid string){
	for k,v:=range workerMgr.resourceInfo{
		if strings.Compare(k,addr)==0{
			for k2,v2:=range v{
				if strings.Compare(v2,aid)==0{
					workerMgr.resourceInfo[k]=append(v[:k2],v[k2+1:]...)
					return
				}
			}
		}
	}
}
func (workerMgr *WorkerMgr) deployUpdateNotify() {
	var (
		getResp            *clientv3.GetResponse
		watchStartRevision int64
		watcher            clientv3.Watcher
		err                error
	)
	if getResp, err = workerMgr.kv.Get(context.TODO(), "/config/pc", clientv3.WithPrefix()); err != nil {
		logging.G_Logger.Warn("同步配置失败")
	}
	for _, ev := range getResp.Kvs {
		keys := strings.Split(string(ev.Key), "/")
		workerMgr.updateConfig(keys[3], keys[4], string(ev.Value))
	}
	watchStartRevision = getResp.Header.Revision + 1
	watcher = clientv3.NewWatcher(workerMgr.client)
	for {
		rch := watcher.Watch(context.Background(), "/config/pc", clientv3.WithPrefix(), clientv3.WithRev(watchStartRevision))
		for wresp := range rch {
			for _, ev := range wresp.Events {
				switch ev.Type {
				case mvccpb.PUT:
					keys := strings.Split(string(ev.Kv.Key), "/")
					workerMgr.updateConfig(keys[3], keys[4], string(ev.Kv.Value))
				}
			}
		}
	}
}

func (workerMgr *WorkerMgr) ResourceUpdateNotify() {
	var (
		getResp            *clientv3.GetResponse
		watchStartRevision int64
		watcher            clientv3.Watcher
		err                error
	)
	if getResp, err = workerMgr.kv.Get(context.TODO(), "/resource", clientv3.WithPrefix()); err != nil {
		logging.G_Logger.Warn("同步配置失败")
	}
	for _, ev := range getResp.Kvs {
		keys := strings.Split(string(ev.Key), "/")
		workerMgr.updateResourceInfo(keys[2], keys[3])
	}
	watchStartRevision = getResp.Header.Revision + 1
	watcher = clientv3.NewWatcher(workerMgr.client)
	for {
		rch := watcher.Watch(context.Background(), "/resource", clientv3.WithPrefix(), clientv3.WithRev(watchStartRevision))
		for wresp := range rch {
			for _, ev := range wresp.Events {
				switch ev.Type {
				case mvccpb.PUT:
					keys := strings.Split(string(ev.Kv.Key), "/")
					workerMgr.updateResourceInfo(keys[2], keys[3])
				case mvccpb.DELETE:
					keys := strings.Split(string(ev.Kv.Key), "/")
					workerMgr.deleteResourceInfo(keys[2],keys[3])
				}
			}
		}
	}
}
func (workerMgr *WorkerMgr) ListWorkers() (workerArr []e.WorkerInfo, err error) {
	var (
		getResp  *clientv3.GetResponse
		kv       *mvccpb.KeyValue
		workerIp string
	)
	workerArr = make([]e.WorkerInfo, 0)
	if getResp, err = workerMgr.kv.Get(context.TODO(), e.JOB_WORKER_DIR, clientv3.WithPrefix()); err != nil {
		return
	}
	for _, kv = range getResp.Kvs {
		if utils2.IsDaemon(string(kv.Key)) || strings.Compare(string(kv.Value), "Daemon") == 0 {
			workerIp = utils2.ExtractWorkerIP(string(kv.Key))
			pcConfig,ok:= workerMgr.clientsInfo[workerIp]
			if !ok{
				pcConfig=&e.ConfigInfo{}
			}
			serverInfo := e.WorkerInfo{
				Ip:     workerIp,
				Server: make([]string, 0),
				Config:*pcConfig,
			}
			serverInfo.Server=append(serverInfo.Server,"Daemon")
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

func (workerMgr *WorkerMgr) isAllowPut(serverName string) bool {
	if len(setting.RunningSetting.IgnoreTopic) > 0 { //判断当前是否有忽略入库的topic
		for _, v := range setting.RunningSetting.IgnoreTopic {
			if m, _ := regexp.MatchString(v, serverName); m {
				return false
			}
		}
	}
	return true
}

// Daemon上线时调用，表示展期机器的上线
func (workerMgr *WorkerMgr) PutServerInfo(serverName string, serverType string) (err error) {
	var (
		topic string
	)
	if !workerMgr.isAllowPut(serverName) {
		logging.G_Logger.Info(fmt.Sprintf("当前推送的topic:" + serverName + ",在忽略名单当中"))
		return
	}
	if len(serverType) > 0 {
		topic = e.JOB_SERVER_DIR + serverName + "/" + serverType
	} else {
		topic = e.JOB_SERVER_DIR + serverName
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	_, err = workerMgr.client.Put(ctx, topic, serverType)
	if err != nil {
		logging.G_Logger.Warn("put  Host info fail:" + err.Error())
	}
	return
}

//获取所有的客户端
func (workerMgr *WorkerMgr) GetAllClient() (clients []string, err error) {
	var (
		resp *clientv3.GetResponse
	)
	clients = make([]string, 0)
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	resp, err = workerMgr.client.Get(ctx, e.JOB_SERVER_DIR, clientv3.WithPrefix())
	if err != nil {
		logging.G_Logger.Warn("get cleitns error:" + err.Error())
		return clients, err
	}
	for _, ev := range resp.Kvs {
		tmp := strings.Split(string(ev.Key), "/")
		clients = append(clients, tmp[2])
	}
	return
}
func (workerMgr *WorkerMgr) GetClientConfigInfo() map[string]*e.ConfigInfo {
	return workerMgr.clientsInfo
}
func (workerMgr *WorkerMgr)GetAllClientInfo(onlineServer []string)map[string]interface{}{
	fmt.Println("GetAllClientInfo")
	data:=make(map[string]interface{})
	tmpOnlineList,err:= workerMgr.ListWorkers()
	tmpOfflineList:=make([]string,0)
	allServer,err:=G_workerMgr.GetAllClient()
	if err==nil{
		for _,v:=range  allServer{
			isOffline:=true
			for _,onlineClient:=range tmpOnlineList{
				if strings.Compare(v,onlineClient.Ip)==0{
					isOffline=false
				}
			}
			if isOffline{
				tmpOfflineList=append(tmpOfflineList,v)
			}
		}
	}
	clientsConfigInfo:= workerMgr.GetClientConfigInfo()
	resourcesInfo:= workerMgr.GetResourceInfo()
	for k,onlineClient:=range tmpOnlineList{
		if info,ok:=clientsConfigInfo[onlineClient.Ip];ok{
			tmpOnlineList[k].Config=*info
		}
		if resource,ok:=resourcesInfo[onlineClient.Ip];ok{
			tmpOnlineList[k].Resource=resource
		}else{
			tmpOnlineList[k].Resource=make([]string,0)
		}
	}
	data["offline"]=tmpOfflineList
	data["online"]=tmpOnlineList
	data["server"]=onlineServer
	//responseBody:=e.ClientResponseInfo{Offline:tmpOfflineList,Online:tmpOnlineList,Server:onlineServer}
	return data
}
func (workerMgr *WorkerMgr) GetResourceInfo()map[string][]string {
	return workerMgr.resourceInfo
}
