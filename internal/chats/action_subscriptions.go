package chats

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (chatHandler *ChatHandler) HandleSubscriptionAction(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatContext := ctx.Value("chatContext").(*ChatContext)

	subscriptions, _ := chatHandler.Options.SubscriptionHandler.GetSubscriptionsFromChat(chatContext.Chat.ID)

	if len(subscriptions) == 0 {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "You have not added any subscription. Subscribe with /subscribe",
		})
	} else {
		output := "Your subscriptions:\n"

		for _, subscription := range subscriptions {
			output += fmt.Sprintf("\n%s", subscription.String())
		}

		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   output,
		})
	}
}
