Daemon能力
========

Daemon支持单机模式，组网模式两种工作模式

提供控制展项计算机的能力

提供控制展项内容的打开，关闭，状态查询等通用能力

可根据命令从指定位置更新展项内容


ZEBus能力
=======

管理所有Daemon（组网，消息路由，更新daemon，安装daemon插件，分发展项内容）

管理展厅所有硬件设备，包含投影，pc机，沙盘，灯光等（打开，关闭，查询状态）

硬件状态定义：在线（可通信，pc机指daemon工作正常），开机，开机中，关机，关机中



GUI业务系统（快展运维人员使用系统）
===================


维护Daemon程序更新

远程控制功能，可以远程控制任意主机（集成VNC）

配置展厅基本硬件信息，包括IP地址，设备类型，通道信息，编号信息等，支持导出到快展系统中，系统间信息数据同步

根据配置生成通用硬件控制**元命令**

支持动态对硬件进行分组，生成**批处理命令**

支持组合元命令，配置延时等功能，**组成宏命令**


**重要交换信息，多系统统一**

硬件设备ID

硬件命令ID （建议 格式：“设备ID@操作码” 如“pc001@open”）

