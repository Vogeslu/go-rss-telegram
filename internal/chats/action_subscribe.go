package chats

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog/log"
	"math"
	"net/url"
	"rss-telegram/internal/utils"
	"slices"
	"strconv"
)

type SubscribeActionStep int

const (
	AskURL SubscribeActionStep = iota
	AskAddPattern
	EnterPattern
)

type SubscribeAction struct {
	step               SubscribeActionStep
	url                *url.URL
	addPattern         bool
	patternSuggestions []string
	pattern            string
	feed               *gofeed.Feed
}

func (chatHandler *ChatHandler) SwitchToSubscribeAction(chatContext *ChatContext) {
	log.Debug().Msgf("Chat %d is switching to subscribe action", chatContext.Chat.ID)

	chatContext.CurrentAction = Subscribe
	chatContext.ActionData = &SubscribeAction{
		step: AskURL,
	}
}

func (chatHandler *ChatHandler) HandleSubscribeActionStart(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Enter a url",
	})
}

func (chatHandler *ChatHandler) HandleSubscribeActionMessage(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatContext := ctx.Value("chatContext").(*ChatContext)
	actionData := chatContext.ActionData.(*SubscribeAction)

	switch actionData.step {
	case AskURL:
		chatHandler.HandleAskUrl(ctx, b, update)
	case AskAddPattern:
		chatHandler.HandleAskAddPattern(ctx, b, update)
	case EnterPattern:
		chatHandler.HandleEnterPattern(ctx, b, update)
	}
}

func (chatHandler *ChatHandler) HandleAskUrl(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatContext := ctx.Value("chatContext").(*ChatContext)
	actionData := chatContext.ActionData.(*SubscribeAction)

	message := update.Message.Text

	parsedUrl, err := url.ParseRequestURI(message)

	if err != nil {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Please enter a valid url",
		})
		return
	}

	fp := gofeed.NewParser()
	actionData.feed, err = fp.ParseURL(parsedUrl.String())
	if err != nil {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Could not receive data from feed, please enter a valid url",
		})
		return
	}

	actionData.url = parsedUrl

	actionData.step = AskAddPattern

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Do you want to add a search pattern?",
		ReplyMarkup: &models.ReplyKeyboardMarkup{
			Keyboard: [][]models.KeyboardButton{
				{
					{Text: "Yes"},
					{Text: "No"},
				},
			},
			OneTimeKeyboard: true,
		},
	})
}

func (chatHandler *ChatHandler) HandleAskAddPattern(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatContext := ctx.Value("chatContext").(*ChatContext)
	actionData := chatContext.ActionData.(*SubscribeAction)

	message := update.Message.Text

	actionData.addPattern = message == "Yes"

	if actionData.addPattern {
		subscriptions, _ := chatHandler.Options.SubscriptionHandler.GetSubscriptionsFromChat(chatContext.Chat.ID)

		var existingPattern []string
		hasSuggestions := false

		optionsText := "You can enter the number of one of the existing pattern:\n"

		for i, sub := range subscriptions {
			if sub.SearchPattern == "" {
				continue
			}

			if slices.Contains(existingPattern, sub.SearchPattern) {
				continue
			}

			existingPattern = append(existingPattern, sub.SearchPattern)
			optionsText += fmt.Sprintf("\n%d - %s", i, sub.SearchPattern)

			hasSuggestions = true
		}

		actionData.patternSuggestions = existingPattern

		var options = make([][]models.KeyboardButton, int(math.Ceil(float64(len(existingPattern))/3)))

		for i, _ := range existingPattern {
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

		if hasSuggestions {
			text := fmt.Sprintf("Enter the pattern (e. g. 'polls' to only receive items with title, url or description containing 'polls')\n\nYou can add multiple words separated by a comma.\n\n%s", optionsText)

			utils.SendChunkedMessage(text, ctx, b, update.Message.Chat.ID, 4000, &models.ReplyKeyboardMarkup{Keyboard: options, OneTimeKeyboard: true})
		} else {
			text := "Enter the pattern (e. g. 'polls' to only receive items with title, url or description containing 'polls')\n\nYou can add multiple words separated by a comma."

			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   text,
			})
		}

		actionData.step = EnterPattern
	} else {
		chatHandler.AddSubscription(ctx, b, update)
	}
}

func (chatHandler *ChatHandler) HandleEnterPattern(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatContext := ctx.Value("chatContext").(*ChatContext)
	actionData := chatContext.ActionData.(*SubscribeAction)

	message := update.Message.Text

	i, err := strconv.Atoi(message)
	if err == nil {
		if i >= 0 && i < len(actionData.patternSuggestions) {
			message = actionData.patternSuggestions[i]
		}
	}

	actionData.pattern = message

	chatHandler.AddSubscription(ctx, b, update)
}

func (chatHandler *ChatHandler) AddSubscription(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatContext := ctx.Value("chatContext").(*ChatContext)
	actionData := chatContext.ActionData.(*SubscribeAction)

	subscription := chatHandler.Options.SubscriptionHandler.NewSubscription(actionData.url, chatContext.Chat.ID, actionData.pattern)

	_, err := chatHandler.Options.SubscriptionHandler.AddSubscription(chatContext.Chat.ID, subscription)
	if err != nil {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Subscription could not be added.",
		})
		return
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Subscribed to %s (%s)", actionData.url.String(), actionData.feed.Title),
	})

	chatHandler.SwitchToCancelAction(chatContext)
}
