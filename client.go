package main

import (
	"encoding/json"
	"fmt"
	"github.com/wenchangshou2/zebus/pkg/certification"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/zap"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wenchangshou2/zebus/pkg/setting"

	"github.com/gorilla/websocket"
	"github.com/wenchangshou2/zebus/pkg/e"
	"github.com/wenchangshou2/zebus/pkg/logging"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 5 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from peer.
	maxMessageSize = 500 * 1024
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  500 * 1024,
	WriteBufferSize: 500 * 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	sync.RWMutex
	hub           *ZEBUSD
	conn          *websocket.Conn
	Ip            string
	Mac           string
	Topic         string
	SocketName    string
	MessageType   int
	IsRegister    bool
	SocketType    string
	send          chan []byte
	sendMessage   chan *Message
	AuthStatus    bool
	client        *clientv3.Client
	kv            clientv3.KV
	lease         clientv3.Lease
	serverType    string
	CancelChannel chan interface{}
	register      *Register
	messageCount  uint64
	messageBytes  uint64
	memoryMsgChan chan *Message
	idFactory     *guidFactory
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}
			if err := w.Close(); err != nil {
				return
			}
		case msg, ok := <-c.memoryMsgChan:
			fmt.Println("msg111111", msg, ok)
			//c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			//if !ok{
			//	c.conn.WriteMessage(websocket.CloseMessage,[]byte{})
			//	return
			//}
			//w,_:=c.conn.NextWriter(websocket.BinaryMessage)
			//w.Write(msg.Body)
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) login(params map[string]interface{}) bool {
	isOk, err := certification.G_Certification.Login(params)
	if err != nil {
		d := make(map[string]interface{})
		d["state"] = 400
		d["message"] = err.Error()
		msg, _ := c.generateResponse(&d)
		c.send <- msg
		return false
	}
	if !isOk {
		d := make(map[string]interface{})
		d["state"] = 400
		d["message"] = "登录失败"
		msg, _ := c.generateResponse(&d)
		c.send <- msg
		c.AuthStatus = false
		return false
	}
	c.AuthStatus = true
	return true
}
func (c *Client) initRegister(socketName string) {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
		err    error
	)
	config = clientv3.Config{
		Endpoints:   []string{setting.EtcdSetting.ConnStr},
		DialTimeout: 5 * time.Second,
	}
	if client, err = clientv3.New(config); err != nil {
		return
	}
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	c.register = &Register{
		client:     client,
		kv:         kv,
		lease:      lease,
		serverType: "Server",
		serverName: socketName,
	}

	go c.register.keepOnline()
	return
}

//注册到Daemon的事件
func (c *Client) registerToDaemon(data e.RequestCmd) {
	arguments := data.Arguments
	if ip, ok := arguments["ip"]; ok {
		c.Ip = ip.(string)
	}
	if mac, ok := arguments["mac"]; ok {
		c.Mac = mac.(string)
	}
	if topic, ok := arguments["topic"]; ok {
		c.Topic = topic.(string)
	}

	if data.SocketType != "Daemon" {
		c.SocketType = "Services"
		c.SocketName = strings.TrimPrefix(data.SocketName, "/")
		if setting.EtcdSetting.Enable {
			G_workerMgr.PutServerInfo(c.SocketName, "Server")
			c.initRegister(c.SocketName)
		}
	} else {
		c.SocketType = "Daemon"
		nameAry := strings.Split(data.SocketName, "/")
		if len(nameAry) <= 2 {
			remoteIp := strings.Split(c.conn.RemoteAddr().String(), ":")[0]
			c.SocketName = fmt.Sprintf("/zebus/%s", remoteIp)
			if setting.EtcdSetting.Enable {
				G_workerMgr.PutServerInfo(remoteIp, "Daemon")
			}
		} else {
			c.SocketName = strings.TrimPrefix(data.SocketName, "/")
		}
	}
	c.Ip = strings.Split(c.conn.RemoteAddr().String(), ":")[0]
	c.IsRegister = true
	// c.send <-
	b, err := c.generateResponse(&map[string]interface{}{
		"MessageType": data.MessageType,
		"Action":      data.MessageType,
	})
	if err == nil {
		if c.send != nil {
			c.send <- b
		}
	}
	c.hub.register <- c
}
func (c *Client) generateResponse(data *map[string]interface{}) ([]byte, error) {
	result := make(map[string]interface{})
	result["message"] = "成功"
	result["state"] = 0
	result["senderName"] = "/zebus"
	for k, v := range *data {
		result[k] = v
	}
	return json.Marshal(result)
}

func (c *Client) execute(data []byte) {
	type zeBusCmd struct {
		Action       string `json:"action"`
		ReceiverName string `json:"receiverName"`
		SenderName   string `json:"sendername"`
	}
	d := make(map[string]interface{})
	cmd := zeBusCmd{}
	err := json.Unmarshal(data, &cmd)
	if err != nil {
		logging.G_Logger.Error("解析json错误:" + err.Error())
		return
	}
	switch cmd.Action {
	case " getClients":
		if setting.EtcdSetting.Enable {
			tmpOnlineList, err := G_workerMgr.ListWorkers()
			tmpOfflineList := make([]string, 0)
			allServer, err := G_workerMgr.GetAllClient()
			if err == nil {
				for _, v := range allServer {
					isOffline := true
					for _, onlineClient := range tmpOnlineList {
						if strings.Compare(v, onlineClient.Ip) == 0 {
							isOffline = false
						}
					}
					if isOffline {
						tmpOfflineList = append(tmpOfflineList, v)
					}
					d["online"] = tmpOnlineList
					d["offline"] = tmpOfflineList
				}
			}
		} else {
			d = c.hub.GetAllClientInfo()
		}
	case "getAuthoricationStatus":
		d["status"] = G_Authorization.Status
		//d=G_workerMgr.ListWorkers()
	}
	if len(cmd.SenderName) == 0 {
		return
	}
	rtu := make(map[string]interface{})
	rtu["state"] = 0
	rtu["message"] = "成功"
	if err != nil {
		rtu["state"] = 400
		rtu["message"] = err.Error()
	}
	rtu["receiverName"] = cmd.SenderName
	rtu["Action"] = cmd.Action
	rtu["senderName"] = "/zebus"
	if len(d) > 0 {
		rtu["data"] = d
	}
	b, err := json.Marshal(rtu)
	if err != nil {
		logging.G_Logger.Error(fmt.Sprintf("send data 错误:%v", err))
		return
	}
	c.hub.forward <- b
}

// 未授权消息
func (c *Client) unAuthorization(recv string) {
	var (
		rtu map[string]interface{}
	)
	rtu = make(map[string]interface{})
	if len(recv) > 0 {
		rtu["Action"] = "UnAuthorization"
		rtu["receiverName"] = recv
		rtu["senderName"] = "/zebus"
		byte, _ := json.Marshal(rtu)
		c.hub.forward <- byte
	}
}
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
		if setting.EtcdSetting.Enable {
			if c.register != nil {
				c.register.CancelChannel <- true
			}
		}
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
				break
			}
			logging.G_Logger.Error("接收失败:" + err.Error())
			break
		}
		logging.G_Logger.Info(string(message), zap.String("type", "message"))
		data := e.RequestCmd{}
		if err = json.Unmarshal(message, &data); err != nil {
			tmp := fmt.Sprintf("解析json错误:%s,错误原因:%s", string(message), err.Error())
			logging.G_Logger.Error(tmp)
			continue
		}
		if strings.Compare(data.MessageType, "RegisterToDaemon") == 0 {
			if setting.ServerSetting.Auth {
				if !c.login(data.Auth) {
					continue
				}
			}
			c.registerToDaemon(data)
			continue
		}
		if !c.IsRegister { //如果当前没有初始不接受任何指令
			continue
		}
		if strings.Compare(data.ReceiverName, "/zebus") == 0 {
			c.execute(message)
		} else {
			c.hub.forward <- message
		}
	}
}
func (c *Client) process(data []byte) {
	if setting.EtcdSetting.Enable {
		G_ScheduleMgr.ProcessData <- data
	} else {
		c.hub.forward <- data
	}
}
func (c *Client) GenerateID() MessageID {
retry:
	id, err := c.idFactory.NewGUID()
	if err != nil {
		time.Sleep(time.Millisecond)
		goto retry
	}
	return id.Hex()
}
func (c *Client) PutMessage(m *Message) error {
	c.RLock()
	defer c.RUnlock()
	err := c.put(m)
	if err != nil {
		return err
	}
	atomic.AddUint64(&c.messageCount, 1)
	atomic.AddUint64(&c.messageBytes, uint64(len(m.Body)))
	return nil
}
func (c *Client) put(m *Message) error {
	fmt.Println("put")
	select {
	case c.memoryMsgChan <- m:
	default:
		//b:=bufferPoolGet()
	}
	return nil
}

// serveWs handles websocket requests from  the peer.
func serveWs(hub *ZEBUSD, w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	// defer conn.Close()
	if err != nil {
		log.Println(err)
		return
	}
	rand.Seed(time.Now().UnixNano())
	client := &Client{
		hub:           hub,
		conn:          conn,
		send:          make(chan []byte, 256),
		idFactory:     NewGUIDFactory(int64(rand.Intn(10000))),
		memoryMsgChan: make(chan *Message, setting.AppSetting.MemQueueSize),
		sendMessage:   make(chan *Message, 0),
	}
	tmp := map[string]interface{}{}
	tmp["ip"] = strings.Split(conn.RemoteAddr().String(), ":")[0]
	tmp["Service"] = "registerCall"
	conn.WriteJSON(tmp)
	time.AfterFunc(5*time.Second, func() {
		if !client.IsRegister {
			close(client.send)
			client.conn.Close()
		}
	})
	go client.writePump()
	go client.readPump()
}
