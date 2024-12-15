package bot

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (botHandler *BotHandler) contextMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		chatContext, err := botHandler.Options.ChatHandler.UpsertChatContext(&update.Message.Chat)
		if err != nil {
			_ = sendMessage(b, ctx, update.Message.Chat.ID, "Your message could not be processed.")
			return
		}

		ctxWithChat := context.WithValue(ctx, "chatContext", chatContext)

		next(ctxWithChat, b, update)
	}
}
