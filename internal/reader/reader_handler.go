package reader

import (
	"context"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"rss-telegram/internal/bot"
	"rss-telegram/internal/subscription"
	"time"
)

type ReaderHandlerOptions struct {
	RedisDb             *redis.Client
	BotHandler          *bot.BotHandler
	SubscriptionHandler *subscription.SubscriptionHandler
	Interval            time.Duration
}

type ReaderHandler struct {
	Options *ReaderHandlerOptions

	TickerIds map[string]uuid.UUID
	Tickers   map[uuid.UUID]*SubscriptionTicker

	Context context.Context
}

func NewReaderHandler(options *ReaderHandlerOptions) *ReaderHandler {
	readerHandler := &ReaderHandler{
		Options:   options,
		TickerIds: make(map[string]uuid.UUID),
		Tickers:   make(map[uuid.UUID]*SubscriptionTicker),
		Context:   context.Background(),
	}

	eventListener := &subscription.ReaderEventListener{
		AddSubscription:    readerHandler.AddSubscription,
		RemoveSubscription: readerHandler.RemoveSubscription,
	}

	readerHandler.Options.SubscriptionHandler.ReaderEventListener = eventListener

	return readerHandler
}

func (readerHandler *ReaderHandler) AddSubscriptions() error {
	subscriptions, err := readerHandler.Options.SubscriptionHandler.GetAllSubscriptions()
	if err != nil {
		return err
	}

	for _, sub := range subscriptions {
		readerHandler.AddSubscription(sub)
	}

	return nil
}

func (readerHandler *ReaderHandler) AddSubscription(subscription *subscription.Subscription) {
	log.Debug().Msgf("Adding subscription %s by %d to reader handler", subscription.URL.String(), subscription.ChatId)

	_, ticker := readerHandler.findExistingTicker(subscription)

	if ticker == nil {
		ticker = readerHandler.NewSubscriptionTicker(subscription)

		id := uuid.New()

		readerHandler.Tickers[id] = ticker
		readerHandler.RunSubscriptionTicker(ticker)
	} else {
		ticker.Subscriptions = append(ticker.Subscriptions, subscription)
	}

}

func (readerHandler *ReaderHandler) RemoveSubscription(subscription *subscription.Subscription) {
	log.Debug().Msgf("Removing subscription %s by %d from reader handler", subscription.URL.String(), subscription.ChatId)

	id, ticker := readerHandler.findExistingTicker(subscription)

	if ticker == nil {
		return
	}

	if len(ticker.Subscriptions) > 1 {
		for i, sub := range ticker.Subscriptions {
			if sub.Id != subscription.Id {
				continue
			}

			ticker.Subscriptions[i] = ticker.Subscriptions[len(ticker.Subscriptions)-1]
			ticker.Subscriptions = ticker.Subscriptions[:len(ticker.Subscriptions)-1]

			break
		}
	} else if len(ticker.Subscriptions) == 0 {
		close(ticker.Quit)

		delete(readerHandler.Tickers, id)
		delete(readerHandler.TickerIds, subscription.URL.String())
	}
}

func (readerHandler *ReaderHandler) findExistingTicker(subscription *subscription.Subscription) (uuid.UUID, *SubscriptionTicker) {
	id, ok := readerHandler.TickerIds[subscription.URL.String()]
	if !ok {
		return uuid.UUID{}, nil
	}

	ticker, ok := readerHandler.Tickers[id]
	if !ok {
		return uuid.UUID{}, nil
	}

	return id, ticker
}

func (readerHandler *ReaderHandler) RunSubscriptions() {
	log.Debug().Msgf("Starting %d tickers", len(readerHandler.Tickers))

	for _, ticker := range readerHandler.Tickers {
		readerHandler.RunSubscriptionTicker(ticker)
	}
}
