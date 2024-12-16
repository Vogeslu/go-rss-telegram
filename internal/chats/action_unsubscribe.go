package chats

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
	"math"
	"rss-telegram/internal/subscription"
	"strconv"
)

type UnsubscribeAction struct {
	options []*subscription.Subscription
}

func (chatHandler *ChatHandler) SwitchToUnsubscribeAction(chatContext *ChatContext) {
	log.Debug().Msgf("Chat %d is switching to unsubscribe action", chatContext.Chat.ID)

	chatContext.CurrentAction = Unsubscribe
	chatContext.ActionData = &UnsubscribeAction{}
}

func (chatHandler *ChatHandler) HandleUnsubscribeActionStart(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatContext := ctx.Value("chatContext").(*ChatContext)
	actionData := chatContext.ActionData.(*UnsubscribeAction)

	subscriptions, _ := chatHandler.Options.SubscriptionHandler.GetSubscriptionsFromChat(chatContext.Chat.ID)

	if len(subscriptions) == 0 {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "You have not added any subscription. Subscribe with /subscribe",
		})

		chatHandler.SwitchToCancelAction(chatContext)
	} else {
		actionData.options = subscriptions

		output := "Enter or select the number you want to unsubscribe:\n"

		for i, sub := range subscriptions {
			output += fmt.Sprintf("\n%d - %s", i, sub.URL.String())
		}

		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        output,
			ReplyMarkup: getReplyMarkup(actionData.options),
		})
	}
}

func (chatHandler *ChatHandler) HandleUnsubscribeActionMessage(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatContext := ctx.Value("chatContext").(*ChatContext)

	subscriptions, _ := chatHandler.Options.SubscriptionHandler.GetSubscriptionsFromChat(chatContext.Chat.ID)

	message := update.Message.Text

	var foundSubscription *subscription.Subscription = nil

	for i, sub := range subscriptions {
		if strconv.Itoa(i) == message {
			foundSubscription = sub
			break
		}
	}

	if foundSubscription == nil {
		output := "Please enter a valid option:\n"

		for i, sub := range subscriptions {
			output += fmt.Sprintf("\n%d - %s", i, sub.URL.String())
		}

		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        output,
			ReplyMarkup: getReplyMarkup(subscriptions),
		})
		return
	}

	chatHandler.Options.SubscriptionHandler.DeleteSubscription(chatContext.Chat.ID, foundSubscription)

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Unsubscribed from %s", foundSubscription.URL),
	})
}

func getReplyMarkup(subscriptions []*subscription.Subscription) *models.ReplyKeyboardMarkup {
	var options = make([][]models.KeyboardButton, int(math.Ceil(float64(len(subscriptions))/3)))

	for i, _ := range subscriptions {
		row := i / 3
		position := i % 3

		button := models.KeyboardButton{
			Text: fmt.Sprintf("%d", i),
		}

		if position == 0 {
			options[row] = []models.KeyboardButton{
				button,
			}
		} else {
			options[row] = append(options[row], button)
		}

	}

	return &models.ReplyKeyboardMarkup{Keyboard: options, OneTimeKeyboard: true}
}
