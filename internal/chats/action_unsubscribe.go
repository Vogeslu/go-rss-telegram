package chats

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
	"rss-telegram/internal/subscription"
)

type UnsubscribeAction struct {
	subscription *subscription.Subscription
}

func (chatHandler *ChatHandler) SwitchToUnsubscribeAction(chatContext *ChatContext) {
	log.Debug().Msgf("Chat %d is switching to unsubscribe action", chatContext.Chat.ID)

	chatContext.CurrentAction = Unsubscribe
	chatContext.ActionData = &UnsubscribeAction{}
}

func (chatHandler *ChatHandler) HandleUnsubscribeActionStart(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatContext := ctx.Value("chatContext").(*ChatContext)

	subscriptions, _ := chatHandler.Options.SubscriptionHandler.GetSubscriptionsFromChat(chatContext.Chat.ID)

	if len(subscriptions) == 0 {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "You have not added any subscription. Subscribe with /subscribe",
		})

		chatHandler.SwitchToCancelAction(chatContext)
	} else {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Select a subscription",
			ReplyMarkup: getReplyMarkup(subscriptions),
		})
	}
}

func (chatHandler *ChatHandler) HandleUnsubscribeActionMessage(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatContext := ctx.Value("chatContext").(*ChatContext)

	subscriptions, _ := chatHandler.Options.SubscriptionHandler.GetSubscriptionsFromChat(chatContext.Chat.ID)

	message := update.Message.Text

	var foundSubscription *subscription.Subscription = nil

	for _, sub := range subscriptions {
		if sub.URL.String() == message {
			foundSubscription = sub
			break
		}
	}

	if foundSubscription == nil {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Please select a valid option",
			ReplyMarkup: getReplyMarkup(subscriptions),
		})
		return
	}

	chatHandler.Options.SubscriptionHandler.DeleteSubscription(chatContext.Chat.ID, foundSubscription)

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Unsubscribed from %s", message),
	})
}

func getReplyMarkup(subscriptions []*subscription.Subscription) *models.ReplyKeyboardMarkup {
	var options = make([][]models.KeyboardButton, len(subscriptions))

	for i, sub := range subscriptions {
		subscriptionUrl := sub.URL.String()

		options[i] = []models.KeyboardButton{
			{
				Text: subscriptionUrl,
			},
		}
	}

	return &models.ReplyKeyboardMarkup{Keyboard: options, OneTimeKeyboard: true}
}
