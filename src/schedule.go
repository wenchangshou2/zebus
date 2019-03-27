package main

import (
	"fmt"
	"net/http"
)

type ServerList struct{

}
func InitSchedume(addr string)(err error){
	hub:=newHub()
	go hub.run()
	http.HandleFunc("/",func(w http.ResponseWriter,r *http.Request){
		fmt.Println("connected")
		serveWs(hub,w,r)
	})
	go http.ListenAndServe(addr,nil)
	return
}