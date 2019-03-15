package main

import "net/http"

type ServerList struct{

}
func InitSchedume(addr string)(err error){
	hub:=newHub()
	go hub.run()
	http.HandleFunc("/ws",func(w http.ResponseWriter,r *http.Request){
		serveWs(hub,w,r)
	})
	go http.ListenAndServe(addr,nil)
	return
}