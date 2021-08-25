package e

import (
	"sync"
	"sync/atomic"
)

type RequestCmd struct {
	MessageType  string                 `json:"messageType"`
	SocketName   string                 `json:"socketName"`
	SocketType   string                 `json:"SocketType"`
	ReceiverName string                 `json:"receiverName"`
	Action       string                 `json:"Action"`
	Arguments    map[string]interface{} `json:"Arguments"`
	Auth         map[string]interface{} `json:"Auth"`
	Proto        string                 `json:"proto"`
	SenderName   string                 `json:"senderName"`
	RegisterName string
}
type ForwardCmd struct {
	Service      string `json:"Service"`
	Action       string `json:"Action"`
	ReceiverName string `json:"receiverName"`
	SenderName   string `json:"senderName"`
	Type         int    `json:"type"`
}
type ClientResponseInfo struct {
	Online  []WorkerInfo `json:"online"`
	Offline []string     `json:"Offline"`
	Server  []string     `json:"server"`
}

// 客户端配置信息
type ConfigInfo struct {
	Volume int32
	Mute   int32
	sync.RWMutex
}

func (cfg ConfigInfo) GetVolume() int {
	volume := atomic.LoadInt32(&cfg.Volume)
	return int(volume)

}
func (cfg ConfigInfo) GetMute() bool {
	mute := atomic.LoadInt32(&cfg.Mute)
	if mute == 1 {
		return true
	} else {
		return false
	}
}
func (cfg *ConfigInfo) SetVolume(volume int) {
	atomic.StoreInt32(&cfg.Volume, int32(volume))
}
func (cfg *ConfigInfo) SetMute(mute bool) {
	if mute {
		atomic.StoreInt32(&cfg.Mute, 1)
	} else {
		atomic.StoreInt32(&cfg.Mute, 0)
	}
}

type WorkerInfo struct {
	Ip       string
	Server   []string
	Config   ConfigInfo
	Resource []string
}

var (
	JOB_WORKER_DIR         = "/zebus/"
	CONFIG_WORKER_DIR      = "/config/"
	JOB_SERVER_DIR         = "/all/"
	SERVER_DIR             = "/servers/"
	MSG_DIR                = "/zebus/"
	JOB_ONLINE_SERVER_DIR  = "/onlineServer/"
	JOB_HISTORY_SERVER_DIR = "/historyServer/"
)

const (
	ERROR_NOT_EXIST                = 10003
	ERROR_ADD_FAIL                 = 10006
	ERROR_AUTH_CHECK_TOKEN_FAIL    = 20001
	ERROR_AUTH_CHECK_TOKEN_TIMEOUT = 20002
	ERROR_AUTH_TOKEN               = 20003
	ERROR_AUTH                     = 20004
)

var ()
