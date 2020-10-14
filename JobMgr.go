package main

import (
	"context"
	"encoding/json"
	"github.com/coreos/etcd/clientv3"
	"github.com/wenchangshou2/zebus/pkg/e"
	"github.com/wenchangshou2/zebus/pkg/setting"
	"time"
)

type JobMgr struct {
	client *clientv3.Client
	kv clientv3.KV
	lease clientv3.Lease
}
var (
	GJobMgr *JobMgr
)
func (jobMgr *JobMgr)SaveJob(job *e.Job)(oldJob *e.Job,err error){
	var (
		jobKey string
		jobValue []byte
		putResp *clientv3.PutResponse
		oldJobObj e.Job
	)
	jobKey=setting.EtcdSetting.DispatchTopic+job.Name
	if jobValue,err=json.Marshal(job);err!=nil{
		return
	}
	if putResp,err=jobMgr.kv.Put(context.TODO(),jobKey,string(jobValue),clientv3.WithPrevKV());err!=nil{
		return
	}

	if putResp.PrevKv!=nil{
		if err=json.Unmarshal(putResp.PrevKv.Value,&oldJob);err!=nil{
			err=nil
			return
		}
		oldJob=&oldJobObj
	}

	return
}
func InitJobMgr()(err error){
	var (
		config clientv3.Config
		client *clientv3.Client
		kv clientv3.KV
		lease clientv3.Lease
	)
	config = clientv3.Config{
		Endpoints:[]string{setting.EtcdSetting.ConnStr},
		DialTimeout:time.Duration(setting.EtcdSetting.Timeout)*time.Millisecond,
	}
	if client,err=clientv3.New(config);err!=nil{
		return
	}
	kv=clientv3.NewKV(client)
	lease=clientv3.NewLease(client)
	GJobMgr =&JobMgr{
		client: client,
		kv:     kv,
		lease:  lease,
	}
	return
}