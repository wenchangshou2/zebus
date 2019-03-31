package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/wenchangshou2/zebus/src/pkg/e"
	"github.com/wenchangshou2/zebus/src/pkg/logging"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
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

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
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
			}
			break
		}
		data := e.RequestCmd{}
		if err = json.Unmarshal(message, &data); err != nil {
			tmp := fmt.Sprintf("解析json错误:%s", string(message))
			logging.G_Logger.Error(tmp)
			continue
		}
		fmt.Println("messageType", data.MessageType, data.SocketName)
		if strings.Compare(data.MessageType, "RegisterToDaemon") == 0 {
			arguments := data.Arguments
			if ip, ok := arguments["ip"]; ok {
				c.Ip = ip.(string)
			}
			if mac, ok := arguments["mac"]; ok {
				c.Mac = mac.(string)
			}
			if topic, ok := arguments["topic"]; ok {
				fmt.Println("toipc", topic.(string))
				c.Topic = topic.(string)

			}
			if data.SocketType != "Daemon" {
				c.SocketName = data.SocketName
			} else {
				nameAry := strings.Split(data.SocketName, "/")
				if len(nameAry) <= 2 {
					fmt.Println("c.conn.RemoteAddr()", c.conn.RemoteAddr().String(), nameAry[1])
					c.SocketName = fmt.Sprintf("/zebus/%s", strings.Split(c.conn.RemoteAddr().String(), ":")[0])
				} else {
					c.SocketName = data.SocketName
				}
			}
			c.hub.register <- c
		} else {
			c.hub.forward <- message
		}

	}
}

// serveWs handles websocket requests from  the peer.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	tmp := map[string]interface{}{}
	tmp["ip"] = strings.Split(conn.RemoteAddr().String(), ":")[0]
	tmp["Service"] = "registerCall"
	conn.WriteJSON(tmp)
	//client.hub.register <- client
	go client.writePump()

	go client.readPump()
}
