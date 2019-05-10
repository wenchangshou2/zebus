package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/wenchangshou2/zebus/src/pkg/setting"

	"github.com/gorilla/websocket"
	"github.com/wenchangshou2/zebus/src/pkg/e"
	"github.com/wenchangshou2/zebus/src/pkg/logging"
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
	hub *Hub

	// The websocket connection.
	conn        *websocket.Conn
	Ip          string
	Mac         string
	Topic       string
	SocketName  string
	MessageType int
	IsRegister  bool
	SocketType  string
	// Buffered channel of outbound messages.
	send chan []byte
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
				// The hub closed the channel.
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
		case <-ticker.C:
			// if strings.Compare(c.SocketName, "Daemon") == 0 {
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
			// }
		}
	}
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
		c.SocketName = data.SocketName
	} else {
		c.SocketType = "Daemon"
		nameAry := strings.Split(data.SocketName, "/")
		if len(nameAry) <= 2 {
			remoteIp:=strings.Split(c.conn.RemoteAddr().String(), ":")[0]
			c.SocketName = fmt.Sprintf("/zebus/%s", remoteIp)
			G_workerMgr.PutServerInfo(remoteIp)
		} else {
			c.SocketName = data.SocketName
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
		c.send <- b
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
		logging.G_Logger.Error("解析json错误")
		return
	}
	switch cmd.Action {
	case "getClients":
		if setting.EtcdSetting.Enable {
			d["online"], err = G_workerMgr.ListWorkers()
		} else {
			d = c.hub.GetAllClientInfo()
		}
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
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
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
			break
		}

		data := e.RequestCmd{}
		if err = json.Unmarshal(message, &data); err != nil {
			tmp := fmt.Sprintf("解析json错误:%s", string(message))
			logging.G_Logger.Error(tmp)
			continue
		}
		if strings.Compare(data.MessageType, "RegisterToDaemon") == 0 {
			c.registerToDaemon(data)
		} else if strings.Compare(data.ReceiverName, "/zebus") == 0 {
			c.execute(message)
		} else {
			c.process(message)
		}
	}
}
func (c *Client) process(data []byte) {
	fmt.Println("process")
	if setting.EtcdSetting.Enable {
		G_ScheduleMgr.ProcessData<-data
	} else {
		c.hub.forward <- data
	}

}

// serveWs handles websocket requests from  the peer.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	// defer conn.Close()
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
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

