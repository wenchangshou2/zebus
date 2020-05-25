package main

import (
	"encoding/json"
	"fmt"
	"github.com/wenchangshou2/zebus/pkg/logging"
	"github.com/wenchangshou2/zebus/pkg/utils"
	"go.uber.org/zap"
	"strings"
	"sync"
)

type ZEBUSD struct {
	// Registered clients.
	clients   map[*Client]bool
	clientMap map[string]*Client
	online    map[string]bool
	offline   map[string]bool
	// Inbound messages from the clients.
	broadcast chan []byte
	// Register requests from the clients.
	register chan *Client
	forward  chan []byte
	// Unregister requests from clients.
	unregister chan *Client
	mux        sync.RWMutex
	logf       *zap.Logger
	onlineServer map[string]bool
}

func newHub(logf *zap.Logger) *ZEBUSD {
	return &ZEBUSD{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		forward:    make(chan []byte),
		online:     make(map[string]bool),
		offline:    make(map[string]bool),
		onlineServer:make(map[string]bool),
		clientMap:  make(map[string]*Client, 0),
		mux:        sync.RWMutex{},
		logf:       logf,
	}
}
func (h *ZEBUSD) GetAllClientInfo() map[string]interface{} {
	h.mux.RLock()
	defer h.mux.RUnlock()
	rtu := make(map[string]interface{})
	onlineClient := make([]string, 0)
	offlineClient := make([]string, 0)
	for k, _ := range h.online {
		if len(k) > 0 {
			onlineClient = append(onlineClient, k)
		}
	}
	for k, _ := range h.offline {
		if len(k) > 0 {
			offlineClient = append(offlineClient, k)
		}
	}
	rtu["online"] = onlineClient
	rtu["offline"] = offlineClient
	return rtu
}
func (h *ZEBUSD) SetClientInfo(ip string, isRegister bool) {
	h.mux.Lock()
	defer h.mux.Unlock()
	if isRegister {
		h.online[ip] = true   //将当前ip添加到在线列表
		delete(h.offline, ip) //删除离线记录
	} else {
		delete(h.online, ip)
		h.offline[ip] = true
	}
}
func (h *ZEBUSD) trimPrefix(topic string) (newTopic string) {
	newTopic = strings.TrimPrefix(topic, "/zebus")
	newTopic = strings.TrimPrefix(newTopic, "/")
	return
}

func (h *ZEBUSD) forwareClientMessage(client *Client, message []byte) {
	select {
	case client.send <- message:
	default:
		close(client.send)
		delete(h.clients, client)
	}
}
func (h *ZEBUSD) addNewServer(serverName string){
	h.mux.Lock()
	defer h.mux.Unlock()
	h.onlineServer[serverName]=true
}
func (h *ZEBUSD)removeServer(serverName string){
	h.mux.Lock()

	if _,ok:=h.onlineServer[serverName];ok{
		delete(h.onlineServer,serverName)
	}
	h.mux.Unlock()

}
func (h *ZEBUSD) getOnlineServer()[]string{
	fmt.Println("getOnlineServer")
	h.mux.RLock()
	defer h.mux.RUnlock()
	onlineClient :=make([]string,0)
	for k,_:=range h.onlineServer{
		onlineClient=append(onlineClient,k)
	}

	onlineClient=append(onlineClient,"zebus")
	return onlineClient
}


// 将消息转发到子节点
func (h *ZEBUSD) forwardProcess(data []byte) {
	var (
		ReceiverName string
		err          error
	)
	cmdBody := make(map[string]interface{})
	if err := json.Unmarshal(data, &cmdBody); err != nil {
		return
	}
	for client := range h.clients { //遍历所有的在线客户端
		if len(client.SocketName) == 0 { //如果没有注册的名称就 不处理
			continue
		}
		ReceiverName,_= cmdBody["receiverName"].(string)
		ReceiverName=h.trimPrefix(ReceiverName)
		if strings.Compare(ReceiverName, client.SocketName) == 0 { //指定 发送第三方服务
			h.forwareClientMessage(client, data)
			return
		}
		isIp, ip := utils.GetIp(ReceiverName)
		if isIp && strings.Compare(client.SocketType, "Daemon") == 0 && strings.Compare(ip, client.Ip) == 0 { //转发给daemon
			cmdBody["receiverName"] = ReceiverName
			data, err = json.Marshal(cmdBody)
			if err == nil {
				h.forwareClientMessage(client, data)
			}
		}
	}
}
func (h *ZEBUSD) getClients(topicName string) *Client {
	h.mux.Lock()
	defer h.mux.Unlock()
	t, ok := h.clientMap[topicName]
	if ok {
		return t
	}
	ip:=utils.FindIp(topicName)
	if len(ip)<=0{
		return nil
	}
	t, ok = h.clientMap["/zebus/"+ip]
	if ok {
		return t
	}
	fmt.Println("ip",ip,"/zebus/"+ip,h.clientMap)

	return nil
}
func (h *ZEBUSD) run() {
	//t := time.NewTicker(24*time.Hour)
	// t:=time.NewTicker(time.Second)
	for {
		select {
		case client := <-h.register:
			logging.G_Logger.Info("new client up",zap.String("event","ServerOnline"),
				zap.String("clientName",client.Ip))
			h.SetClientInfo(client.Ip, true)
			h.clients[client] = true
			h.clientMap[client.SocketName] = client
			if strings.Compare(client.SocketType,"Services")==0{
				h.addNewServer(client.SocketName)
			}
			//client.SocketName
		case client := <-h.unregister:
			logging.G_Logger.Info("client down",zap.String("event","ServerDropped"),zap.String("clientName",client.Ip))
			if _, ok := h.clients[client]; ok {
				fmt.Println("ok", client.send)
				delete(h.clients, client)
				delete(h.clientMap, client.SocketName)
				close(client.send)
			}
			h.SetClientInfo(client.Ip, false)
			if strings.Compare(client.SocketType,"Services")==0{
				h.removeServer(client.SocketName)
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		case message := <-h.forward:
			h.forwardProcess(message)
		// case <-t.C:
		// 	logging.G_Logger.Info("check license")
		// 	err:=utils.CheckLicense()
		// 	if err!=nil{
		// 		logging.G_Logger.Error("license 到期:"+err.Error())
		// 		panic("License 到期")
		// 	}

		}
	}
}
