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
	MSG_DIR           = "/zebus/"
)
