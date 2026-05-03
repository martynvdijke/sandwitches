package utils

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const flashKey = "flash_messages"

type FlashMessage struct {
	Type    string
	Message string
}

func AddFlash(c *gin.Context, msgType, message string) {
	session := sessions.Default(c)
	var messages []FlashMessage
	if raw := session.Get(flashKey); raw != nil {
		if msgs, ok := raw.([]FlashMessage); ok {
			messages = msgs
		}
	}
	messages = append(messages, FlashMessage{Type: msgType, Message: message})
	session.Set(flashKey, messages)
	session.Save()
}

func GetFlashes(c *gin.Context) []FlashMessage {
	session := sessions.Default(c)
	var messages []FlashMessage
	if raw := session.Get(flashKey); raw != nil {
		if msgs, ok := raw.([]FlashMessage); ok {
			messages = msgs
		}
	}
	session.Delete(flashKey)
	session.Save()
	return messages
}
