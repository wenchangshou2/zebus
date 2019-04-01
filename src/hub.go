package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/wenchangshou2/zebus/src/pkg/e"
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
			fmt.Println("k", k)

			onlineClient = append(onlineClient, k)
		}
	}
	for k, _ := range h.offline {
		if len(k) > 0 {
			fmt.Println("k", k)
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
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			fmt.Println("register")
			h.SetClientInfo(client.Ip, true)
			h.clients[client] = true
		case client := <-h.unregister:

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
			cmdBody := e.ForwardCmd{}
			json.Unmarshal(message, &cmdBody)
			for client := range h.clients {
				fmt.Println("clientl.socket",cmdBody.ReceiverName,client.SocketName)
				if len(client.SocketName)==0{
					continue
				}
				if strings.Compare(client.SocketName, cmdBody.ReceiverName) == 0 || strings.HasPrefix(cmdBody.ReceiverName, client.SocketName) {
					fmt.Println("send")
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}

		}
	}
}
