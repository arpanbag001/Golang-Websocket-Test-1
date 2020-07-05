package config

import "time"

const (
	// WriteWait specifies time allowed to write a message to the client
	WriteWait = 10 * time.Second

	// PongWait specifies time allowed to read the next pong message from the client
	PongWait = 60 * time.Second

	// PingPeriod specifies the time period to send pings to client. Must be less than PongWait.
	PingPeriod = (PongWait * 9) / 10

	// MaxMessageSize specifies maximum message size allowed from client
	MaxMessageSize = 512
)
