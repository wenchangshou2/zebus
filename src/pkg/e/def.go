package e
type RequestCmd struct{
	MessageType string `json:"messageType"`
	SocketName string `json:"socketName"`
	SocketType string `json:"SocketType"`
	Arguments map[string]interface{} `json:"Arguments"`

}
type ForwardCmd struct{
	Service string `json:"Service"`
	Action string `json:"Action"`
	ReceiverName string `json:"receiverName"`
	SenderName string `json:"senderName"`
	Type int `json:"type"`
}