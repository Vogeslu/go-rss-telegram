package bot

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
	"rss-telegram/internal/chats"
)

func (botHandler *BotHandler) registerCommands() {
	log.Debug().Msg("Registering telegram bot commands")

	botHandler.Bot.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, botHandler.startHandler, botHandler.contextMiddleware)

	botHandler.Bot.RegisterHandler(bot.HandlerTypeMessageText, "/cancel", bot.MatchTypeExact, botHandler.cancelHandler, botHandler.contextMiddleware)

	botHandler.Bot.RegisterHandler(bot.HandlerTypeMessageText, "/subscriptions", bot.MatchTypeExact, botHandler.subscriptionHandler, botHandler.contextMiddleware)
	botHandler.Bot.RegisterHandler(bot.HandlerTypeMessageText, "/subscribe", bot.MatchTypeExact, botHandler.subscribeHandler, botHandler.contextMiddleware)
	botHandler.Bot.RegisterHandler(bot.HandlerTypeMessageText, "/unsubscribe", bot.MatchTypeExact, botHandler.unsubscribeHandler, botHandler.contextMiddleware)
}

func (botHandler *BotHandler) startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Hello %s\n\n/subscribe = Subscribe to a new feed\n/unsubscribe = Unsubscribe from an feed\n/subscriptions = Get active subscriptions", update.Message.Chat.Username),
	})
}

func (botHandler *BotHandler) cancelHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatContext := ctx.Value("chatContext").(*chats.ChatContext)

	botHandler.Options.ChatHandler.SwitchToCancelAction(chatContext)
	botHandler.Options.ChatHandler.HandleCancelActionStart(ctx, b, update)
}

func (botHandler *BotHandler) subscriptionHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	botHandler.Options.ChatHandler.HandleSubscriptionAction(ctx, b, update)
}

func (botHandler *BotHandler) subscribeHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatContext := ctx.Value("chatContext").(*chats.ChatContext)

	botHandler.Options.ChatHandler.SwitchToSubscribeAction(chatContext)
	botHandler.Options.ChatHandler.HandleSubscribeActionStart(ctx, b, update)
}

func (botHandler *BotHandler) unsubscribeHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatContext := ctx.Value("chatContext").(*chats.ChatContext)

	botHandler.Options.ChatHandler.SwitchToUnsubscribeAction(chatContext)
	botHandler.Options.ChatHandler.HandleUnsubscribeActionStart(ctx, b, update)
}
