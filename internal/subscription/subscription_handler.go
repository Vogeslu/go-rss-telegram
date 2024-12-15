package subscription

import (
	"context"
	"github.com/redis/go-redis/v9"
	"sync"
)

type SubscriptionHandlerOptions struct {
	RedisDb *redis.Client
}

type SubscriptionHandler struct {
	Options            *SubscriptionHandlerOptions
	Context            context.Context
	subscriptionsCache map[string]*Subscription
	lock               sync.Mutex

	ReaderEventListener *ReaderEventListener
}

type ReaderEventListener struct {
	AddSubscription    func(subscription *Subscription)
	RemoveSubscription func(subscription *Subscription)
}

func NewSubscriptionHandler(options *SubscriptionHandlerOptions) *SubscriptionHandler {
	subscriptionHandler := &SubscriptionHandler{
		Options:            options,
		Context:            context.Background(),
		subscriptionsCache: make(map[string]*Subscription),
	}

	return subscriptionHandler
}
