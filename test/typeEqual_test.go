package test

import (
	"github.com/wenchangshou2/zebus/pkg/utils"
	"strings"
	"testing"
)

func TestEqualDaemonType(t *testing.T) {
	var (
		b bool
	)
	str := "/zebus/192.168.10.2"
	b = utils.IsDaemon(str)
	if !b {
		t.Error("判断错误")
	}
	str = "/zebus/192.168.10.2/resource"
	b = utils.IsDaemon(str)
	if b {
		t.Error("判断错误")
	}
}
func TestGetip(t *testing.T) {
	topic := "192.168.10.27/test"
	strArr := strings.Split(topic, "/")
	t.Log(strArr[0])

}
