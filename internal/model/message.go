package model

type Message struct {
	Type       MsgType `json:"type"`
	Title      string  `json:"title"`
	Data       string  `json:"data"`
	SenderID   int64   `json:"senderID"`
	ReceiverID int64   `json:"receiverID"`
	Time       string  `json:"time"`
}

type MsgType int

const (
	Follow MsgType = iota
	Like
	Event
)
