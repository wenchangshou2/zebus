package main
type Options struct{
	EtcdEnable bool `long:"etcd-enable" description:"是否启动etcd"`
	EtcdServer string `long:"etcd-server" description:"etcd 连接地址"`
	EtcdTimeout int `long:"etcd-timeout" description:"etcd连接超时时间"`
	EtcdDispatch bool `long:"etcd-dispatch" description:"etcd服务调度"`
	ServerBindAddress string `long:"server-bind-address" description:"服务绑定端口"`
	ServerAddress string `long:"server-address" description:"下发的地址"`
}
//type Options struct{
//	EtcdEnable bool `flag:"etcd-enable"`
//	EtcdServer string `flag:"etcd-server"`
//	EtcdTimeout int `flag:"etcd-timeout"`
//	EtcdDispatch bool `flag:"etcd-dispatch"`
//	ServerBindAddress string `flag:"server-bind-address"`
//	ServerAddress string `flag:"server-address"`
//}
//
//func NewOptions()*Options{
//	return &Options{
//		EtcdEnable: true,
//		EtcdServer:"127.0.0.1:2379",
//		EtcdTimeout: 5000,
//		EtcdDispatch: false,
//		ServerBindAddress: "0.0.0.0:8181",
//		ServerAddress: "127.0.0.1",
//	}
//}
//
//func SyncFlagSet(opts *Options)*flag.FlagSet{
//	flagSet:=flag.NewFlagSet("sync",flag.ExitOnError)
//	flagSet.Bool("version",false,"print version string")
//	flagSet.Bool("etcd-enable",true,"是否启用etcd服务")
//	flagSet.String("etcd-server","","etcd 连接字符串")
//	flagSet.Int("etcd-timeout",5000,"etcd 超时连接设置")
//	flagSet.Bool("etcd-dispatch",false,"是否启动调度")
//	flagSet.String("server-bind-address","0.0.0.0:8181","server 绑定字符串")
//	flagSet.String("server-address","127.0.0.1","服务器地址")
//	return flagSet
//
//}

