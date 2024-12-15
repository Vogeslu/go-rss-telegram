package reader

import (
	"context"
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
	Tickers []*SubscriptionTicker
	Context context.Context
}

func NewReaderHandler(options *ReaderHandlerOptions) *ReaderHandler {
	readerHandler := &ReaderHandler{
		Options: options,
		Context: context.Background(),
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

	readerHandler.Tickers = make([]*SubscriptionTicker, len(subscriptions))

	for i, sub := range subscriptions {
		readerHandler.Tickers[i] = readerHandler.NewSubscriptionTicker(sub)
	}

	return nil
}

func (readerHandler *ReaderHandler) AddSubscription(subscription *subscription.Subscription) {
	log.Debug().Msgf("Adding subscription %s by %d to reader handler", subscription.URL.String(), subscription.ChatId)

	ticker := readerHandler.NewSubscriptionTicker(subscription)

	readerHandler.Tickers = append(readerHandler.Tickers, ticker)
	readerHandler.RunSubscriptionTicker(ticker)
}

func (readerHandler *ReaderHandler) RemoveSubscription(subscription *subscription.Subscription) {
	log.Debug().Msgf("Removing subscription %s by %d from reader handler", subscription.URL.String(), subscription.ChatId)

	for i, ticker := range readerHandler.Tickers {
		if ticker.Subscription.Id != subscription.Id {
			continue
		}

		close(ticker.Quit)

		readerHandler.Tickers[i] = readerHandler.Tickers[len(readerHandler.Tickers)-1]
		readerHandler.Tickers = readerHandler.Tickers[:len(readerHandler.Tickers)-1]

		break
	}
}

func (readerHandler *ReaderHandler) RunSubscriptions() {
	log.Debug().Msgf("Starting %d tickers", len(readerHandler.Tickers))

	for _, ticker := range readerHandler.Tickers {
		readerHandler.RunSubscriptionTicker(ticker)
	}
}
