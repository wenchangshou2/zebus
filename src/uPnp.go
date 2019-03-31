package main

import (
	"encoding/json"
	"net"

	"github.com/wenchangshou2/zebus/src/pkg/setting"
)

const maxBufferSize = 1024

func GetServerInfo() (data []byte, err error) {
	tmp := map[string]interface{}{}
	tmp["Service"] = "RegisterInfo"
	tmp["ip"] = setting.ServerSetting.ServerIp
	tmp["port"] = setting.ServerSetting.ServerPort
	return json.Marshal(tmp)
}
func upnpServer(conn net.PacketConn) {
	buffer := make([]byte, maxBufferSize)
	for {
		_, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			continue
		}
		data, err := GetServerInfo()
		if err == nil {
			conn.WriteTo(data, addr)
		}
	}
}
func InituPnpServer(ip string, port int) (err error) {
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(ip),
	}
	pc, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return
	}
	go upnpServer(pc)
	return
}