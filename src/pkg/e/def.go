package e

type RequestCmd struct {
	MessageType  string                 `json:"messageType"`
	SocketName   string                 `json:"socketName"`
	SocketType   string                 `json:"SocketType"`
	ReceiverName string                 `json:"receiverName"`
	Action       string                 `json:"Action"`
	Arguments    map[string]interface{} `json:"Arguments"`
}
type ForwardCmd struct {
	Service      string `json:"Service"`
	Action       string `json:"Action"`
	ReceiverName string `json:"receiverName"`
	SenderName   string `json:"senderName"`
	Type         int    `json:"type"`
}
type WorkerInfo struct {
	Ip     string
	Server []string
}

var (
	JOB_WORKER_DIR    = "/zebus/"
	CONFIG_WORKER_DIR = "/config/"
	JOG_SERVER_DIR = "/server/"
	MSG_DIR           = "/zebus/"
)

const(
	SUCCESS =200
	ERROR = 500
	INVALID_PARAMS =400
	ERROR_UPLOAD_SAVE_PACKAGE_FAIL=30001
	ERROR_UPLOAD_CHECK_PACKAGE_FAIL=30002
	ERROR_UPLOAD_CHECK_PACKAGE_FORMAT = 30003
	ERROR_JSON_PARSER_ERROR = 40001
	ERROR_ADD_CLIENT_FAIL=40002
	ERROR_ADD_TASK_FAIL=40003
	ERROR_UPDATE_TASK_DOWNLOAD_STATUS_FAIL=40004
	ERROR_UPDATE_TASK_DOWNLOAD_SCHEDULE_FAIL=40005
	ERROR_ADD_SERVER_FAIL=40006
	ERROR_GET_SERVER_FAIL=40007
	ERROR_DELETE_CLIENT_FAIL=40008
	ERROR_NOT_EXIST=10003
	ERROR_ADD_FAIL=10006
	ERROR_AUTH_CHECK_TOKEN_FAIL    = 20001
	ERROR_AUTH_CHECK_TOKEN_TIMEOUT = 20002
	ERROR_AUTH_TOKEN               = 20003
	ERROR_AUTH                     = 20004

)
var (
	
)