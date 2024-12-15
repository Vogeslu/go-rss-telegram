package chats

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type CurrentAction int

const (
	None CurrentAction = iota
	Subscribe
	Unsubscribe
)

func (chatHandler *ChatHandler) PassMessageHandlerToAction(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatContext := ctx.Value("chatContext").(*ChatContext)

	switch chatContext.CurrentAction {
	case Subscribe:
		chatHandler.HandleSubscribeActionMessage(ctx, b, update)
	case Unsubscribe:
		chatHandler.HandleUnsubscribeActionMessage(ctx, b, update)
	default:
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Enter a command to start an action (e.g. /subscription)",
		})
	}
}
