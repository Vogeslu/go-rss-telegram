package reader

import (
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/mmcdole/gofeed"
	"rss-telegram/internal/subscription"
	"sync"
	"time"
)

type SubscriptionTicker struct {
	Subscription *subscription.Subscription
	Ticker       *time.Ticker
	Quit         chan struct{}

	InRequest bool

	Parser        *gofeed.Parser
	FailedFetches int

	Lock sync.Mutex
}

func (readerHandler *ReaderHandler) NewSubscriptionTicker(subscription *subscription.Subscription) *SubscriptionTicker {
	subscriptionTicker := &SubscriptionTicker{
		Subscription:  subscription,
		Ticker:        time.NewTicker(readerHandler.Options.Interval),
		Quit:          make(chan struct{}),
		InRequest:     false,
		Parser:        gofeed.NewParser(),
		FailedFetches: 0,
	}

	return subscriptionTicker
}

func (readerHandler *ReaderHandler) RunSubscriptionTicker(subscriptionTicker *SubscriptionTicker) {
	go func() {
		for {
			select {
			case <-subscriptionTicker.Ticker.C:
				if subscriptionTicker.InRequest {
					return
				}

				subscriptionTicker.Lock.Lock()
				subscriptionTicker.InRequest = true
				subscriptionTicker.Lock.Unlock()

				feed, err := subscriptionTicker.Parser.ParseURL(subscriptionTicker.Subscription.URL.String())

				if err != nil {
					subscriptionTicker.Lock.Lock()
					subscriptionTicker.FailedFetches++
					subscriptionTicker.Lock.Unlock()

					if subscriptionTicker.FailedFetches >= 3 {
						_, _ = readerHandler.Options.BotHandler.Bot.SendMessage(readerHandler.Options.BotHandler.Options.Context, &bot.SendMessageParams{
							ChatID: subscriptionTicker.Subscription.ChatId,
							Text:   fmt.Sprintf("Could not fetch feed from %s for three times, unsubscribing...", subscriptionTicker.Subscription.URL.String()),
						})

						readerHandler.Options.SubscriptionHandler.DeleteSubscription(subscriptionTicker.Subscription.ChatId, subscriptionTicker.Subscription)
						return
					}
				} else {
					_ = readerHandler.handleFeed(subscriptionTicker, feed)
				}

				subscriptionTicker.Lock.Lock()
				subscriptionTicker.InRequest = false
				subscriptionTicker.Lock.Unlock()
			case <-subscriptionTicker.Quit:
				subscriptionTicker.Ticker.Stop()
				return
			}
		}
	}()
}
