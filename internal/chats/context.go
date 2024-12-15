package chats

import (
	"github.com/go-telegram/bot/models"
)

type ChatContext struct {
	Chat          *models.Chat `json:"chat"`
	CurrentAction CurrentAction
	ActionData    interface{}
}

func (chatHandler *ChatHandler) UpsertChatContext(chat *models.Chat) (*ChatContext, error) {
	chatContext, ok := chatHandler.chatContextCache[chat.ID]
	if ok {
		return chatContext, nil
	}

	chatContext = chatHandler.newChatContext(chat)

	chatHandler.lock.Lock()
	chatHandler.chatContextCache[chat.ID] = chatContext
	chatHandler.lock.Unlock()

	return chatContext, nil
}

func (chatHandler *ChatHandler) newChatContext(chat *models.Chat) *ChatContext {
	return &ChatContext{
		Chat:          chat,
		CurrentAction: None,
	}
}
