package main

import (
	"encoding/json"
	"fmt"
	"regexp"
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
	fmt.Println("checkIp")
	arr1 := strings.Split(source, "/")
	arr2 := strings.Split(target, "/")
	fmt.Println("arr", len(arr1), len(arr2))
	if len(arr1) > 2 && len(arr2) > 2 {
		ip1Str := arr1[2]
		ip2Str := arr2[2]
		fmt.Println("ip1111", ip1Str, ip2Str)
		if h.isIp(ip1Str) && h.isIp(ip2Str) {
			if strings.Compare(ip1Str, ip2Str) != 0 {
				return false
			}

		}
	}
	return true
}
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			fmt.Println("register", client.SocketName)
			h.SetClientInfo(client.Ip, true)
			h.clients[client] = true
		case client := <-h.unregister:
			fmt.Println("unregister", client.SocketName)
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
				if len(client.SocketName) == 0 {
					continue
				}
				if strings.Compare(client.SocketName, cmdBody.ReceiverName) == 0 || strings.HasPrefix(cmdBody.ReceiverName, client.SocketName) && h.checkIp(client.SocketName, cmdBody.ReceiverName) {
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
