package utils

import (
	"github.com/wenchangshou2/zebus/pkg/e"
	"strings"
)

func ExtractWorkerIP(regKey string) string {
	return strings.TrimPrefix(regKey, e.JOB_WORKER_DIR)
}

//func IsIP(ip string) (b bool) {
//	if m, _ := regexp.MatchString("^[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}$", ip); !m {
//		return false
//	}
//	return true
//}
func IsDaemon(regKey string) bool {
	var (
		arr []string
	)
	arr = strings.Split(regKey, "/")
	if strings.Compare(arr[1], "zebus") != 0 {
		return false
	}
	if len(arr) > 3 {
		return false
	}
	if !IsIp(arr[2]) {
		return false
	}
	return true

}
func ExtractServerName(regKey string) (string, string) {
	var (
		arr []string
	)
	arr = strings.Split(regKey, "/")
	return arr[2], arr[3]
}
