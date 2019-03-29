package main

import (
	"encoding/json"
	"fmt"
	"github.com/wenchangshou2/zebus/src/pkg/e"
	"strings"
)

type Hub struct{
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client
	forward chan []byte
	// Unregister requests from clients.
	unregister chan *Client
}
func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		forward:make(chan  []byte),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			fmt.Println("register",string(client.SocketName))
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			fmt.Println("unregister")
		case message := <-h.broadcast:
			fmt.Println("brodcast")
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		case message:=<-h.forward://
		fmt.Println("forward",string(message),h.clients)
			cmdBody:=e.ForwardCmd{}
			json.Unmarshal(message,&cmdBody)
			//tmp:=strings.Split(cmdBody.ReceiverName)
			for client:=range h.clients{
				//fmt.Println("name",client.SocketName,cmdBody.ReceiverName)
				//if strings .Compare(client.SocketName,cmdBody.ReceiverName)==0{
				//	fmt.Println("yyyy1222",client.SocketName)
				//	go func() {
				//		client.send<-message
				//	}()
				//}
				fmt.Println(client.SocketName,cmdBody.ReceiverName,strings.Compare(client.SocketName,cmdBody.ReceiverName))
				if strings.Compare(client.SocketName,cmdBody.ReceiverName)==0||strings.HasPrefix(cmdBody.ReceiverName,client.SocketName){
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