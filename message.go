package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
    "errors"
)

const (
	MsgIDLength       = 16
	minValidMsgLength = MsgIDLength + 8 + 2
)

type MessageID [MsgIDLength]byte

type Message struct {
	ID         MessageID
	Body       []byte
	Timestamp  int64
	Attempts   uint16
	Topic      []byte
	deliveryTS time.Time
	clientID   int64
	pri        int64
	index      int
	deferred   time.Duration
}

func NewMessage(id MessageID, body []byte, topic []byte) *Message {
	return &Message{
		ID:        id,
		Body:      body,
		Timestamp: time.Now().UnixNano() / 1000000,
		Topic:     topic,
	}
}
func (m *Message) WriteTo(w io.Writer) (int64, error) {
	var buf [12]byte
	var total int64
	binary.BigEndian.PutUint64(buf[:8], uint64(m.Timestamp))
	binary.BigEndian.PutUint16(buf[8:10], m.Attempts)
	binary.BigEndian.PutUint16(buf[10:12], uint16(len(m.Topic)))
	n, err := w.Write(buf[:])
	total += int64(n)
	if err != nil {
		return total, err
	}
	// 写入topic
	n, err = w.Write(m.Topic[:])
	total += int64(n)
	if err != nil {
		return total, err
	}
	// 写入一个ID，用来标识当前的消息ID
	n, err = w.Write(m.ID[:])
	total += int64(n)
	if err != nil {
		return total, err
	}
	// 写入消息body
	n, err = w.Write(m.Body)
	total += int64(n)
	if err != nil {
		return total, err
	}
	return total, err
}
func decodeMessage(b []byte) (*Message, error) {
	var msg Message
	var topicLen uint16
    var msgLen uint16
    msgLen=uint16(len(b))
	if len(b) < minValidMsgLength {
		return nil, fmt.Errorf("invalid message buff size (%d)", len(b))
	}
	msg.Timestamp = int64(binary.BigEndian.Uint64(b[:8]))
	fmt.Println("timestamp", msg.Timestamp)
	msg.Attempts = binary.BigEndian.Uint16(b[8:10])
	topicLen = binary.BigEndian.Uint16(b[10:12])
	msg.Topic = make([]byte, topicLen)
    if (12+topicLen)>msgLen{
        return nil,errors.New("解析topic失败")
    }
	copy(msg.Topic[:], b[12:12+topicLen])
    if msgLen<(12+topicLen+MsgIDLength){
        return nil,errors.New("解析id失败")
    }
	copy(msg.ID[:], b[12+topicLen:12+topicLen+MsgIDLength])
	msg.Body = b[12+topicLen+MsgIDLength:]
	fmt.Println("topic", string(msg.Topic), b[10:12], topicLen, string(b[12+topicLen:]))
	return &msg, nil
}
