package main

import (
	"encoding/json"
	"net"

	"github.com/wenchangshou2/zebus/pkg/setting"
)

const maxBufferSize = 1024

func GetServerInfo() (data []byte, err error) {
	tmp := map[string]interface{}{}
	tmp["Service"] = "RegisterInfo"
	tmp["ip"] = setting.ServerSetting.ServerIP
	tmp["port"] = setting.ServerSetting.ServerPort
	tmp["etcd_enable"]=setting.EtcdSetting.Enable
	tmp["etcd_address"]=setting.EtcdSetting.ConnStr
	return json.Marshal(tmp)
}

// 处理服务发现的事件
func upnpServer(conn net.PacketConn) {
	buffer := make([]byte, maxBufferSize)
	for {
		_, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			continue
		}
		data, err := GetServerInfo()
		if err == nil {
			_, _ = conn.WriteTo(data, addr)
		}
	}
}
func InitUPNPServer(ip string, port int) (err error) {
	var (
		pc *net.UDPConn
	)
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(ip),
	}
	pc, err = net.ListenUDP("udp", &addr)
	if err != nil {
		return
	}
	go upnpServer(pc)
	return
}
