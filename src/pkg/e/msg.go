
package e

var MsgFlags=map[int]string{
	SUCCESS:                         "ok",
	ERROR:                           "fail",
	INVALID_PARAMS:                  "请求参数错误",
	ERROR_UPLOAD_CHECK_PACKAGE_FAIL: "检测升级包失败",
	ERROR_UPLOAD_CHECK_PACKAGE_FORMAT: "校验上传文件错误，包格式或大小有问题",
	ERROR_UPLOAD_SAVE_PACKAGE_FAIL:"保存升级包失败",
	ERROR_JSON_PARSER_ERROR:"JSON解析错误",
	ERROR_ADD_CLIENT_FAIL:"添加客户端失败",
	ERROR_ADD_TASK_FAIL:"添加任务失败",
	ERROR_DELETE_CLIENT_FAIL:"删除客户端失败",
}
func GetMsg(code int)string{
	msg,ok:=MsgFlags[code]
	if ok{
		return msg
	}
	return MsgFlags[ERROR]
}