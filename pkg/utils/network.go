package utils

import (
	"regexp"
	"strings"
)

//获取当前topic的ip
func GetIp(topic string) (bool, string) {
	var (
		arr []string
	)
	arr = strings.Split(topic, "/")
	isIp := IsIp(arr[0])
	return isIp, arr[0]
}
func IsIp(str string) bool {
	matched, _ := regexp.MatchString("(2(5[0-5]{1}|[0-4]\\d{1})|[0-1]?\\d{1,2})(\\.(2(5[0-5]{1}|[0-4]\\d{1})|[0-1]?\\d{1,2})){3}", str)
	return matched
}
func CheckIp(source, target string) bool {
	arr1 := strings.Split(source, "/")
	arr2 := strings.Split(target, "/")
	if len(arr1) > 2 && len(arr2) > 2 {
		ip1Str := arr1[2]
		ip2Str := arr2[2]
		if IsIp(ip1Str) && IsIp(ip2Str) {
			if strings.Compare(ip1Str, ip2Str) != 0 {
				return false
			}
		}
	}
	return true
}
//提取ip
func FindIp(input string) string {
	partIp := "(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])"
	grammer := partIp+"\\."+partIp+"\\."+partIp+"\\."+partIp
	matchMe := regexp.MustCompile(grammer)
	return matchMe.FindString(input)
}