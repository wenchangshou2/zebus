package utils

import (
	"fmt"
	"testing"
)

func TestGetCpuID(t *testing.T) {
	cpuId, err := getCpuId()
	if err != nil {
		t.Error("获取CPU ID失败")
		return
	}
	if len(cpuId) <= 0 {
		t.Error("获取CPU ID错误")
		return
	}
	t.Logf("当前获取到的CPUID:%s", cpuId)
}
func TestGetMachineUUID(t *testing.T) {
	hostId, err := getMachineUUID()
	if err != nil {
		t.Error("获取machine id失败")
		return
	}
	if len(hostId) <= 0 {
		t.Error("获取machine id无效")
		return
	}
	t.Logf("获取的machine id:%s", hostId)
}
func TestGetUUID(t *testing.T) {
	uuid, err := GetSystemUUID()
	fmt.Println("get system uuid", uuid, err)
}
