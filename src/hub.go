package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/wenchangshou2/zebus/src/pkg/logging"
)

type Hub struct {
	// Registered clients.
	clients map[*Client]bool
	online  map[string]bool
	offline map[string]bool
	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client
	forward  chan []byte
	// Unregister requests from clients.
	unregister chan *Client
	mux        sync.RWMutex
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		forward:    make(chan []byte),
		online:     make(map[string]bool),
		offline:    make(map[string]bool),
		mux:        sync.RWMutex{},
	}
}
func (h *Hub) GetAllClientInfo() map[string]interface{} {
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
func (h *Hub) SetClientInfo(ip string, isRegister bool) {
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
func (h *Hub) isIp(str string) bool {
	matched, _ := regexp.MatchString("(2(5[0-5]{1}|[0-4]\\d{1})|[0-1]?\\d{1,2})(\\.(2(5[0-5]{1}|[0-4]\\d{1})|[0-1]?\\d{1,2})){3}", str)
	return matched
}
func (h *Hub) checkIp(source, target string) bool {
	arr1 := strings.Split(source, "/")
	arr2 := strings.Split(target, "/")
	if len(arr1) > 2 && len(arr2) > 2 {
		ip1Str := arr1[2]
		ip2Str := arr2[2]
		if h.isIp(ip1Str) && h.isIp(ip2Str) {
			if strings.Compare(ip1Str, ip2Str) != 0 {
				return false
			}

		}
	}
	return true
}
func (h *Hub) trimPrefix(topic string) (newTopic string) {
	newTopic = strings.TrimPrefix(topic, "/zebus")
	newTopic = strings.TrimPrefix(newTopic, "/")
	return

}

//获取当前topic的ip
func (h *Hub) getIp(topic string) (bool, string) {
	var (
		arr []string
	)
	arr = strings.Split(topic, "/")
	isIp := h.isIp(arr[0])
	return isIp, arr[0]
}
func (h *Hub) forwareClientMessage(client *Client, message []byte) {
	select {
	case client.send <- message:
	default:
		close(client.send)
		delete(h.clients, client)
	}
}
func (h *Hub) forwardProcess(data []byte) {
	var (
		ReceiverNmae string
		ok           bool
		err          error
	)
	cmdBody := make(map[string]interface{})
	if err := json.Unmarshal(data, &cmdBody); err != nil {
		return
	}
	if ReceiverNmae, ok = cmdBody["receiverName"].(string); !ok {
		logging.G_Logger.Info(fmt.Sprintf("当前的消息没有接收者,直接抛弃:%s",string(data)))
		return
	}

	for client := range h.clients { //遍历所有的在线客户端
		if len(client.SocketName) == 0 { //如果没有注册的名称就 不处理
			continue
		}
		ReceiverNmae = h.trimPrefix(ReceiverNmae)
		if strings.Compare(ReceiverNmae, "/dm") == 0 || strings.Compare(ReceiverNmae, "dm") == 0 {
			logging.G_Logger.Debug("receiver message:" + string(data))
		}
		if strings.Compare(ReceiverNmae, client.SocketName) == 0 { //指定 发送第三方服务
			h.forwareClientMessage(client, data)
			return
		}
		isIp, ip := h.getIp(ReceiverNmae)
		if isIp && strings.Compare(client.SocketType, "Daemon") == 0 && strings.Compare(ip, client.Ip) == 0 { //转发给daemon
			cmdBody["receiverName"] = ReceiverNmae
			data, err = json.Marshal(cmdBody)
			if err == nil {
				h.forwareClientMessage(client, data)
			}
		}

		//if strings.Compare(client.SocketName, cmdBody.ReceiverName) == 0 || strings.HasPrefix(cmdBody.ReceiverName, client.SocketName) && h.checkIp(client.SocketName, cmdBody.ReceiverName) {
		//	select {
		//	case client.send <- data:
		//	default:
		//		close(client.send)
		//		delete(h.clients, client)
		//	}
		//}
	}

}
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			logging.G_Logger.Info(fmt.Sprintf("register:%s", client.SocketName))
			h.SetClientInfo(client.Ip, true)
			h.clients[client] = true
		case client := <-h.unregister:
			logging.G_Logger.Info("unregister " + client.SocketName)
			if _, ok := h.clients[client]; ok {
				fmt.Println("ok", client.send)
				delete(h.clients, client)
				close(client.send)
			}
			h.SetClientInfo(client.Ip, false)

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		case message := <-h.forward: //

			h.forwardProcess(message)
		}
	}
}
