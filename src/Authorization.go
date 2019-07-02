package main

import (
	"time"
)

var private_key string=
	`MIICWwIBAAKBgQCqsRHp2WYgrkI1qrS+7T/vWvI4z4sUNwKVsXnMaDnY60c11923
nRM/FHyYAF5qF6KlbQ0aqf8DBz8bDTsyUHb+jrNu+SAoQjVO26AIvAVhogK5qLUz
6D6xomngyo93zjH1wO4ptc8x02Mlumju6YLKNWpKIR5/MRj7vYz8FtZ8bwIDAQAB
AoGALfLgqZzWOzHtrNi5MzRWo65NyjFEdTqhvX47FWVxPQ2I69uiWc004yQ2rgxb
Xh/irrl+b5EXjs8ik7uqFc9HWKqV1O9kNpAS6qla1yxUPLCIBGpxcErrk6GnPpxp
eU4Se5X41n1A/bGtINks29n6YhAVxdiUMFMVlGp9ARjesmECQQDaf5RX5Ne+w/vM
S9Bo7GzZuf81aTgILur1a5U7vXnABeTXiAODALhgcwoMmT7JN/8dRhAwk0XRImot
m33zfcGpAkEAx/zztz4SfXFXwegMdLaO2N/NIjZhj9kBFBg1KH0bsWwI5Sfcr27d
Azi94GZ4N+IkAoXTv9DkWhooCd8oNO/MVwJAHaK6QyWl4ZkBeRc7YE/Y/7sLk3n/
AJUkhz8dUaoEbngeLuGi4EzjtSlFTqomavJuZtEO9xeym4gYcLErZzBCaQJAN6Hn
TkdHL3wzNG7P4DvUmwIO94B3PWPZh/R//SZoaL+r7ctb+bV2Z+oF8AGxWaJf8A+4
avi6PVJfZvecILXAewJAIVfQIZPH3BmKqwfPNn9Y7J8+o5Uc6b4Brk/5VNyWHcWK
sOgeRICIbqubBO3vXmNeaJPDV5B28sVnSTgWqf0Wdg==
`
var public_key string=`
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCqsRHp2WYgrkI1qrS+7T/vWvI4
z4sUNwKVsXnMaDnY60c11923nRM/FHyYAF5qF6KlbQ0aqf8DBz8bDTsyUHb+jrNu
+SAoQjVO26AIvAVhogK5qLUz6D6xomngyo93zjH1wO4ptc8x02Mlumju6YLKNWpK
IR5/MRj7vYz8FtZ8bwIDAQAB
`
type AuthorizationInfo struct{
	Uuid string //系统唯一标识
	Type int //检验类型
	ExpireDate time.Time //到期时间
	IsVerify bool //是否定期校验
	VerifyCycle time.Time //检验周期
}
type AuthorizationRequest struct{
	Uuid string
}
//func GenerateRequestInfo()string{
//	var (
//		uuid string
//		err error
//	)
//	uuid,err=utils.GetSystemUUID()
//
//}
func InitAuthorization(done chan   bool){
	done <-true
}
