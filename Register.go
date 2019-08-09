package main

import (
	"context"
	"strings"
	"time"

	"github.com/wenchangshou2/zebus/pkg/e"
	"go.etcd.io/etcd/clientv3"
)

type Register struct {
	client        *clientv3.Client
	kv            clientv3.KV
	lease         clientv3.Lease
	localIP       string
	serverType    string
	serverName    string
	CancelChannel chan interface{}
}

func (register *Register) keepOnline() {
	var (
		regKey               string
		leaseGrantResp       *clientv3.LeaseGrantResponse
		err                  error
		keepAliveChan        <-chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp        *clientv3.LeaseKeepAliveResponse
		cancelCtx            context.Context
		cancelFunc           context.CancelFunc
		cancelKeepOnlineCtx  context.Context
		cancelKeepOnlineFunc context.CancelFunc
	)
	register.CancelChannel = make(chan interface{})
	for {

		if strings.Compare(register.serverType, "Server") == 0 {
			regKey = e.SERVER_DIR + register.serverName
		}
		cancelFunc = nil

		if leaseGrantResp, err = register.lease.Grant(context.TODO(), 10); err != nil {
			goto RETRY
		}
		cancelKeepOnlineCtx, cancelKeepOnlineFunc = context.WithCancel(context.TODO())
		if keepAliveChan, err = register.lease.KeepAlive(cancelKeepOnlineCtx, leaseGrantResp.ID); err != nil {
			goto RETRY
		}
		cancelCtx, cancelFunc = context.WithCancel(context.TODO())
		if _, err = register.kv.Put(cancelCtx, regKey, register.serverType, clientv3.WithLease(leaseGrantResp.ID)); err != nil {
			goto RETRY
		}
		for {
			select {
			case keepAliveResp = <-keepAliveChan:
				if keepAliveResp == nil {
					goto RETRY
				}
			case <-register.CancelChannel:
				time.Sleep(1 * time.Second)
				if cancelKeepOnlineFunc != nil {
					cancelKeepOnlineFunc()
				}
				return
			}
		}
	RETRY:
		time.Sleep(1 * time.Second)
		if cancelFunc != nil {
			cancelFunc()
		}
	}
}
