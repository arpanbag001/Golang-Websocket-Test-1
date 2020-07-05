package models

//Message represents a single message of the chat service
type Message struct {
	ID          string `json:"Id"`
	SenderID    string `json:"senderId"`
	RecipientID string `json:"recipientId"`
	Content     string `json:"content"`
	Time        int64  `json:"time"`
}
