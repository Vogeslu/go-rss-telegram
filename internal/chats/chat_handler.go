package chats

import (
	"context"
	"github.com/redis/go-redis/v9"
	"rss-telegram/internal/subscription"
	"sync"
)

type ChatHandlerOptions struct {
	RedisDb             *redis.Client
	SubscriptionHandler *subscription.SubscriptionHandler
}

type ChatHandler struct {
	Options          *ChatHandlerOptions
	Context          context.Context
	chatContextCache map[int64]*ChatContext
	lock             sync.Mutex
}

func NewChatHandler(options *ChatHandlerOptions) *ChatHandler {
	chatHandler := &ChatHandler{
		Options:          options,
		Context:          context.Background(),
		chatContextCache: make(map[int64]*ChatContext),
	}

	return chatHandler
}
