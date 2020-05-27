package main

import "flag"

type Options struct{
	EtcdEnable bool `flag:"etcd-enable"`
	EtcdConnStr string `flag:"etcd-connect"`
	EtcdTimeout int `flag:"etcd-timeout"`
	EtcdDispatch bool `flag:"etcd-dispatch"`
	ServerBindAddress string `flag:"server-bind-address"`
	ServerAddress string `flag:"server-address"`
}

func NewOptions()*Options{
	return &Options{
		EtcdEnable: true,
		EtcdConnStr:"127.0.0.1:2379",
		EtcdTimeout: 5000,
		EtcdDispatch: false,
		ServerBindAddress: "0.0.0.0:8181",
		ServerAddress: "127.0.0.1",
	}
}

func SyncFlagSet(opts *Options)*flag.FlagSet{
	flagSet:=flag.NewFlagSet("sync",flag.ExitOnError)
	flagSet.Bool("version",false,"print version string")
	flagSet.Bool("etcd-enable",true,"是否启用etcd服务")
	flagSet.String("etcd-connect","","etcd 连接字符串")
	flagSet.Int("etcd-timeout",5000,"etcd 超时连接设置")
	flagSet.Bool("etcd-dispatch",false,"是否启动调度")
	flagSet.String("server-bind-address","0.0.0.0:8181","server 绑定字符串")
	flagSet.String("server-address","127.0.0.1","服务器地址")
	return flagSet

}