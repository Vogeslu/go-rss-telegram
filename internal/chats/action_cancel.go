package chats

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

func (chatHandler *ChatHandler) SwitchToCancelAction(chatContext *ChatContext) {
	log.Debug().Msgf("Chat %d cancels action", chatContext.Chat.ID)

	chatContext.CurrentAction = None
	chatContext.ActionData = nil
}

func (chatHandler *ChatHandler) HandleCancelActionStart(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Previous action canceled",
	})
}
