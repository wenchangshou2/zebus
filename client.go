package main

import (
	"bytes"
	"container/heap"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/wenchangshou2/zebus/pkg/certification"
	"github.com/wenchangshou2/zebus/pkg/pqueue"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/zap"
	"log"
	"math"
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
	hub                  *ZEBUSD
	conn                 *websocket.Conn
	Ip                   string
	Mac                  string
	Topic                string
	Group                string //用于分组
	SocketName           string
	MessageType          int
	IsRegister           bool
	SocketType           string
	send                 chan []byte
	sendMessage          chan *Message
	AuthStatus           bool
	client               *clientv3.Client
	kv                   clientv3.KV
	lease                clientv3.Lease
	serverType           string
	CancelChannel        chan interface{}
	register             *Register
	messageCount         uint64
	messageBytes         uint64
	memoryMsgChan        chan *Message
	idFactory            *guidFactory
	WaitRecvMessageMutex sync.Mutex
	WaitRecvMessage      map[MessageID]chan *[]byte //回复的队列
	deferredMessage      map[MessageID]*pqueue.Item
	deferredPQ           pqueue.PriorityQueue
	deferredMutex        sync.Mutex
	proto                string
	exitChan chan bool
}

func (c *Client) AddNewWaitMessage(id MessageID) chan *[]byte {
	c.WaitRecvMessageMutex.Lock()
	chanBody := make(chan *[]byte)
	c.WaitRecvMessage[id] = chanBody
	c.WaitRecvMessageMutex.Unlock()
	return chanBody
}
func (c *Client) DeleteWitMessage(id MessageID) {
	c.WaitRecvMessageMutex.Lock()
	delete(c.WaitRecvMessage, id)
	c.WaitRecvMessageMutex.Unlock()
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
			if msg.deferred != 0 {
				fmt.Println("迟延发送")
				c.PutMessageDeferred(msg, msg.deferred)
				continue
			}
			var buf = &bytes.Buffer{}
			_, err := msg.WriteTo(buf)
			if err != nil {
				logging.G_Logger.Error(fmt.Sprintf("解析memoryMsgChan错误:%v", err))
				continue
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}
			w.Write(buf.Bytes())
			w.Write(newline)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.exitChan:
			return
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
		exit: make(chan bool),
	}
	go c.register.keepOnline()
	return
}

// 初始化消息队列
func (c *Client) InitPQ() {
	pqSize := int(math.Max(1, 100))
	// 初始化回复消息
	c.WaitRecvMessageMutex.Lock()
	c.WaitRecvMessage = make(map[MessageID]chan *[]byte)
	c.WaitRecvMessageMutex.Unlock()

	//初始化延迟消息
	c.deferredMutex.Lock()
	c.deferredMessage = make(map[MessageID]*pqueue.Item)
	c.deferredPQ = pqueue.New(pqSize)
	c.deferredMutex.Unlock()
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
	if group, ok := arguments["group"]; ok {
		c.Group = group.(string)
	} else {
		c.Group = "default"
	}
	c.proto = data.Proto
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
func (c *Client) PutMessageDeferred(msg *Message, timeout time.Duration) {
	atomic.AddUint64(&c.messageCount, 1)
	c.StartDeferredTimeout(msg, timeout)
}
func (c *Client) pushDeferredMessage(item *pqueue.Item) error {
	c.deferredMutex.Lock()
	id := item.Value.(*Message).ID
	_, ok := c.deferredMessage[id]
	if ok {
		c.deferredMutex.Unlock()
		return errors.New("ID 已经存在")
	}
	c.deferredMessage[id] = item
	c.deferredMutex.Unlock()
	return nil
}

func (c *Client) StartDeferredTimeout(msg *Message, timeout time.Duration) error {
	absTs := time.Now().Add(timeout).UnixNano()
	item := &pqueue.Item{Value: msg, Priority: absTs}
	err := c.pushDeferredMessage(item)
	if err != nil {
		return err
	}
	c.addToDeferredPQ(item)
	return nil
}

//添加到队列
func (c *Client) addToDeferredPQ(item *pqueue.Item) {
	c.deferredMutex.Lock()
	heap.Push(&c.deferredPQ, item)
	c.deferredMutex.Unlock()
}

func (c *Client) execute(data []byte) {
	type zeBusCmd struct {
		Action       string `json:"action"`
		ReceiverName string `json:"receiverName"`
		SenderName   string `json:"senderName"`
		GroupName string `json:"group_name"`
		Body string `json:"body"`
	}
	d := make(map[string]interface{})
	cmd := zeBusCmd{}
	err := json.Unmarshal(data, &cmd)
	if err != nil {
		logging.G_Logger.Error("解析json错误:" + err.Error())
		return
	}
	fmt.Println("action",cmd.Action)
	switch cmd.Action {
	case "getClients":
		if setting.EtcdSetting.Enable {
			d = G_workerMgr.GetAllClientInfo(c.hub.getOnlineServer())
		} else {
			d = c.hub.GetAllClientInfo()
		}
	case "getAuthoricationStatus":
		d["status"] = G_Authorization.Status
		//d=G_workerMgr.ListWorkers()
	case "syncServiceInfo":
		c.syncServiceInfo(data)
	case "forwardGroupMessage":
		if len(cmd.GroupName)==0{
			logging.G_Logger.Info("forwardGroupMessage not found group_name")
			return
		}
		c.ForwardGroupMessage(cmd.GroupName,cmd.Body)
	case "ping":
		d["result"]="pong"
	}
	if len(cmd.SenderName) == 0 {
		return
	}
	rtu := make(map[string]interface{})
	rtu["state"] = 0
	rtu["message"] = "成功"
	rtu["receiverName"] = cmd.SenderName
	rtu["Action"] = cmd.Action
	rtu["senderName"] = "/zebus"
	if err != nil {
		rtu["state"] = 400
		rtu["message"] = err.Error()
	}
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
func (c *Client) syncServiceInfo(data []byte) {

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
func (c *Client) checkFrontCondition(data e.RequestCmd) bool {
	// 如果当前的结点没有注册，不
	if !c.IsRegister {
		fmt.Println("未注册")
		return false
	}
	if len(data.ReceiverName) <= 0 {
		return false
	}
	return true
}
func (c *Client) TextMessageProcess(message []byte) (err error) {
	data := e.RequestCmd{}
	if err = json.Unmarshal(message, &data); err != nil {
		tmp := fmt.Sprintf("解析json错误:%s,错误原因:%s", string(message), err.Error())
		logging.G_Logger.Error(tmp)
		return
	}
	if strings.Compare(data.MessageType, "RegisterToDaemon") == 0 {
		if setting.ServerSetting.Auth {
			if !c.login(data.Auth) {
				return
			}
		}
		c.registerToDaemon(data)
		return
	}
	if !c.checkFrontCondition(data) { //未满足条件，直接废弃
		logging.G_Logger.Info("废弃消息", zap.String("type", "MessageObsolete"),
			zap.String("msg", string(message)),
			zap.String("ReceiverName", data.ReceiverName),
			zap.String("SenderName", data.SenderName),
		)
	} else {
		logging.G_Logger.Info("消息成功接收", zap.String("type", "MessageReceiver"),
			zap.String("ReceiverName", data.ReceiverName),
			zap.String("SenderName", data.SenderName),
			zap.String("msg", string(message)))
		if strings.Compare(data.ReceiverName, "/zebus") == 0 {
			c.execute(message)
		} else {
			c.hub.forward <- message
		}
	}
	return nil
}
func (c *Client) BinaryMessageProcess(message []byte) {
	msg, _ := decodeMessage(message)
	if msg==nil||msg.Topic==nil{
		return
	}
	j,_:=json.Marshal(msg)
	fmt.Println("json=",string(j))
	if string(msg.Topic) != "/zebus" && string(msg.Topic) != "zebus" {
		//c.put()
	}
	if c, ok := c.WaitRecvMessage[msg.ID]; ok {
		c <- &msg.Body
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
		close(c.exitChan)
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
		messageType, messageBody, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
				break
			}
			logging.G_Logger.Error("接收失败:" + err.Error())
			break
		}
		if messageType == websocket.BinaryMessage {
			c.BinaryMessageProcess(messageBody)
		} else {
			c.TextMessageProcess(messageBody)
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
	if strings.Compare(c.proto, "text") == 0 {
		select {
		case c.send <- m.Body:
		default:

		}
	} else {
		select {
		case c.memoryMsgChan <- m:
		default:
			fmt.Println("put22")
			//b:=bufferPoolGet()
		}

	}
	return nil
}
func (c *Client) popDeferredMessage(id MessageID) (*pqueue.Item, error) {
	c.deferredMutex.Lock()
	item, ok := c.deferredMessage[id]
	if !ok {
		c.deferredMutex.Unlock()
		return nil, errors.New("ID 没有迟延")
	}
	delete(c.deferredMessage, id)
	c.deferredMutex.Unlock()
	return item, nil
}
func (c *Client) ProcessDeferredQueue(t int64) bool {
	dirty := false
	for {
		c.deferredMutex.Lock()
		item, _ := c.deferredPQ.PeekAndShift(t)
		c.deferredMutex.Unlock()

		if item == nil {
			goto exit
		}
		dirty = true
		msg := item.Value.(*Message)
		_, err := c.popDeferredMessage(msg.ID)
		if err != nil {
			goto exit
		}
		msg.deferred = 0
		fmt.Println("延迟", string(msg.Body))
		c.send <- msg.Body
		//c.put(msg)
	}
exit:
	return dirty
}
func (c *Client) loop() {
	t:=time.NewTicker(5*time.Second)
	for {
		select {
			case <-c.exitChan:
				return
			case <-t.C:
				now := time.Now().UnixNano()
				c.ProcessDeferredQueue(now)
		}
	}
}

func (c *Client) ForwardGroupMessage(groupName string,body string) {
	data:=e.ForWardGroupMessage{}
	data.Body=[]byte(body)
	data.GroupName=groupName
	go func() {
		c.hub.forwardGroupMessage<-data
	}()

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
		proto:         "text",
		exitChan: make(chan bool),
	}
	tmp := map[string]interface{}{}
	tmp["ip"] = strings.Split(conn.RemoteAddr().String(), ":")[0]
	tmp["Service"] = "registerCall"
	conn.WriteJSON(tmp)
	client.InitPQ()
	go client.writePump()
	go client.readPump()
	go client.loop()
}
