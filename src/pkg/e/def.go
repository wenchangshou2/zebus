package e
type RequestCmd struct{
	MessageType string `json:"messageType"`
	SocketName string `json:"socketName"`
	SocketType string `json:"SocketType"`
}