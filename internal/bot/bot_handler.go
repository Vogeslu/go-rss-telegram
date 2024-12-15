package bot

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/redis/go-redis/v9"
	"rss-telegram/internal/chats"
	"rss-telegram/internal/subscription"
)

type BotHandlerOptions struct {
	BotToken            string
	RedisDb             *redis.Client
	ChatHandler         *chats.ChatHandler
	SubscriptionHandler *subscription.SubscriptionHandler
	Context             context.Context
}

type BotHandler struct {
	Options *BotHandlerOptions
	Bot     *bot.Bot
}

func NewBotHandler(options *BotHandlerOptions) (*BotHandler, error) {
	botHandler := &BotHandler{
		Options: options,
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(botHandler.contextMiddleware(botHandler.handler)),
	}

	b, err := bot.New(options.BotToken, opts...)
	if err != nil {
		return nil, err
	}

	botHandler.Bot = b

	botHandler.registerCommands()

	return botHandler, nil
}

func (botHandler *BotHandler) handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	botHandler.Options.ChatHandler.PassMessageHandlerToAction(ctx, b, update)
}
