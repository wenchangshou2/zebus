package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/wenchangshou2/zebus/pkg/e"
	"github.com/wenchangshou2/zebus/pkg/logging"
	"github.com/wenchangshou2/zebus/pkg/utils"
	"go.uber.org/zap"
)

//ZEBUSD zebus hub
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
	group    map[string][]*Client
	// Unregister requests from clients.
	unregister          chan *Client
	mux                 sync.RWMutex
	logf                *zap.Logger
	onlineServer        map[string]bool
	forwardGroupMessage chan e.ForWardGroupMessage
}

func newHub(logf *zap.Logger) *ZEBUSD {
	return &ZEBUSD{
		broadcast:           make(chan []byte),
		register:            make(chan *Client),
		unregister:          make(chan *Client),
		clients:             make(map[*Client]bool),
		forward:             make(chan []byte, 32),
		online:              make(map[string]bool),
		offline:             make(map[string]bool),
		onlineServer:        make(map[string]bool),
		clientMap:           make(map[string]*Client),
		group:               make(map[string][]*Client),
		forwardGroupMessage: make(chan e.ForWardGroupMessage),
		mux:                 sync.RWMutex{},
		logf:                logf,
	}
}

// GetAllClientInfo: 获取所有客户端信息
func (h *ZEBUSD) GetAllClientInfo() map[string]interface{} {
	h.mux.RLock()
	defer h.mux.RUnlock()
	rtu := make(map[string]interface{})
	onlineClient := make([]string, 0)
	offlineClient := make([]string, 0)
	for k := range h.online {
		if len(k) > 0 {
			onlineClient = append(onlineClient, k)
		}
	}
	for k := range h.offline {
		if len(k) > 0 {
			offlineClient = append(offlineClient, k)
		}
	}
	rtu["online"] = onlineClient
	rtu["offline"] = offlineClient
	return rtu
}

// SetClientInfo: 设置客户端信息
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

// trimPrefix: 移除前缀
func (h *ZEBUSD) trimPrefix(topic string) (newTopic string) {
	newTopic = strings.TrimPrefix(topic, "/zebus")
	newTopic = strings.TrimPrefix(newTopic, "/")
	return
}

// forwardClientMessage: 转发消息到客户端
func (h *ZEBUSD) forwardClientMessage(client *Client, message []byte) {
	defer func() {
		if err := recover(); err != nil {
			logging.G_Logger.Error(fmt.Sprintf("forwardClientMessage 程序异常退出:%s,转发的topic:%s,转发的内容:%s", err, client.Topic, string(message)))
		}
	}()
	select {
	case client.send <- message:
	default:
		close(client.send)
		delete(h.clients, client)
	}
}

// addNewServer 添加新的服务
func (h *ZEBUSD) addNewServer(serverName string) {
	h.mux.Lock()
	defer h.mux.Unlock()
	h.onlineServer[serverName] = true
}

// @title removeServer
// @param serverName string 服务名称
func (h *ZEBUSD) removeServer(serverName string) {
	h.mux.Lock()
	// if _, ok := h.onlineServer[serverName]; ok {
	delete(h.onlineServer, serverName)
	// }
	h.mux.Unlock()

}

// getOnlineServer: 获取在线服务
func (h *ZEBUSD) getOnlineServer() []string {
	h.mux.RLock()
	defer h.mux.RUnlock()
	onlineClient := make([]string, 0)
	for k := range h.onlineServer {
		onlineClient = append(onlineClient, k)
	}
	onlineClient = append(onlineClient, "zebus")
	return onlineClient
}

// forwardProcess:将消息转发到子节点
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
		ReceiverName, _ = cmdBody["receiverName"].(string)
		ReceiverName = h.trimPrefix(ReceiverName)
		if strings.Compare(ReceiverName, client.SocketName) == 0 { //指定 发送第三方服务
			h.forwardClientMessage(client, data)
			return
		}
		isIp, ip := utils.GetIp(ReceiverName)
		if isIp && strings.Compare(client.SocketType, "Daemon") == 0 && strings.Compare(ip, client.Ip) == 0 { //转发给daemon
			cmdBody["receiverName"] = ReceiverName
			data, err = json.Marshal(cmdBody)
			if err == nil {
				logging.G_Logger.Debug("转发消息:" + string(data))
				h.forwardClientMessage(client, data)
			}
		}
	}
}

// getClients: 获取客户端
func (h *ZEBUSD) getClients(topicName string) *Client {
	var (
		client *Client
		ok     bool
	)
	h.mux.Lock()
	defer h.mux.Unlock()
	if client, ok = h.clientMap[topicName]; ok {
		return client
	}
	ip := utils.FindIp(topicName)
	if len(ip) <= 0 {
		return nil
	}
	if client, ok = h.clientMap["/zebus/"+ip]; ok {
		return client
	}
	return nil
}
func (h *ZEBUSD) run() {
	for {
		select {
		case client := <-h.register:
			logging.G_Logger.Debug("new client up", zap.String("event", "ServerOnline"),
				zap.String("clientName", client.Ip))
			h.SetClientInfo(client.Ip, true)
			h.clients[client] = true
			h.clientMap[client.SocketName] = client
			h.addGroupMember(client)
			if strings.Compare(client.SocketType, "Services") == 0 {
				h.addNewServer(client.SocketName)
			}
		case client := <-h.unregister:
			logging.G_Logger.Debug("client down", zap.String("event", "ServerDropped"), zap.String("clientName", client.Ip))
			h.removeGroupMember(client)
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.clientMap, client.SocketName)
				close(client.send)
			}
			h.SetClientInfo(client.Ip, false)
			if strings.Compare(client.SocketType, "Services") == 0 {
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
		case message := <-h.forwardGroupMessage:
			h.forwardGroupMessageProcess(message)
		}
	}
}

// forwardGroupMessageProcess 转发消息到特定分组
func (h *ZEBUSD) forwardGroupMessageProcess(message e.ForWardGroupMessage) {
	if g, ok := h.group[message.GroupName]; ok {
		for _, c := range g {
			logging.G_Logger.Info(fmt.Sprintf("forward GroupMessage:topic(%s)", c.SocketName))
			h.forwardClientMessage(c, message.Body)
		}
	}
}

// addGroupMember 添加分组成员
func (h *ZEBUSD) addGroupMember(c *Client) {
	logging.G_Logger.Debug(fmt.Sprintf("add group member,%s", c.Group))
	if g, ok := h.group[c.Group]; ok {
		g = append(g, c)
		h.group[c.Group] = g
	} else {
		g = make([]*Client, 0)
		g = append(g, c)
		h.group[c.Group] = g
	}
}

// removeGroupMember 移除分组成员
func (h *ZEBUSD) removeGroupMember(c *Client) {
	if g, ok := h.group[c.Group]; ok {
		for k, _c := range g {
			if _c == c {
				g = append(g[:k], g[k+1:]...)
				h.group[c.Group] = g
				return
			}
		}
	}
}
