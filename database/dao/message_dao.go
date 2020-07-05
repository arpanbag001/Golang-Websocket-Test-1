package dao

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"websocket_test_1/models"
)

// CreateMessage creates a message with the provided details and returns the created message
func CreateMessage(senderID, recipientID, content string) (*models.Message, error) {

	//Generate the messageId
	b := make([]byte, 4)
	rand.Read(b)
	messageID := hex.EncodeToString(b)

	//Create the message
	newMessage := models.Message{
		ID:          "mockMessage" + messageID,
		SenderID:    senderID,
		RecipientID: recipientID,
		Content:     content,
		Time:        time.Now().Unix(),
	}

	//TODO: Insert to database

	//Return the newly created message
	return &newMessage, nil
}
