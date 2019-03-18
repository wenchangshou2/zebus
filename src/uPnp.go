package main

import (
	"encoding/json"
	"fmt"
	"github.com/wenchangshou2/zebus/src/pkg/setting"
	"net"
)
const maxBufferSize=1024
func GetServerInfo()(data []byte,err error){
	tmp:=map[string]interface{}{}
	tmp["Service"]="RegisterInfo"
	tmp["ip"]=setting.ServerSetting.ServerIp
	tmp["port"]=setting.ServerSetting.ServerPort
	return json.Marshal(tmp)
	//tmp["ip"]=setting.AppSetting
}
func upnpServer(conn  net.PacketConn){
	buffer:=make([]byte,maxBufferSize)
	for{
		_,addr,err:=conn.ReadFrom(buffer)
		if err!=nil{
			continue
		}
		fmt.Println("buffer",string(buffer))
		data,err:=GetServerInfo()
		if err==nil{
			conn.WriteTo(data,addr)
		}
	}
}
func InituPnpServer(ip string,port int)(err error){
	fmt.Println("initpnp")
	addr:=net.UDPAddr{
		Port:port,
		IP:net.ParseIP(ip),
	}
	pc,err:=net.ListenUDP("udp",&addr)
	fmt.Println("pc",pc,addr)
	if err!=nil{
		fmt.Println("err",err)
		return
	}
	go upnpServer(pc)
	return
}