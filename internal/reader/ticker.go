package reader

import (
	"errors"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog/log"
	"net/url"
	"rss-telegram/internal/subscription"
	"sync"
	"time"
)

type SubscriptionTicker struct {
	URL           *url.URL
	Subscriptions []*subscription.Subscription
	Ticker        *time.Ticker
	Quit          chan struct{}

	InRequest bool

	Parser        *gofeed.Parser
	FailedFetches int
	WaitTimeout   *time.Time

	Lock sync.Mutex
}

func (readerHandler *ReaderHandler) NewSubscriptionTicker(sub *subscription.Subscription) *SubscriptionTicker {
	log.Info().Msgf("Instantiating new subscription ticker for %s by %d", sub.URL.String(), sub.ChatId)

	subscriptionTicker := &SubscriptionTicker{
		URL:           sub.URL,
		Subscriptions: []*subscription.Subscription{sub},
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
				if subscriptionTicker.WaitTimeout != nil {
					if subscriptionTicker.WaitTimeout.After(time.Now()) {
						return
					} else {
						subscriptionTicker.Lock.Lock()
						subscriptionTicker.WaitTimeout = nil
						subscriptionTicker.Lock.Unlock()
					}
				}

				if subscriptionTicker.InRequest {
					return
				}

				log.Trace().Msgf("Requesting %s's feed", subscriptionTicker.URL.String())

				subscriptionTicker.Lock.Lock()
				subscriptionTicker.InRequest = true
				subscriptionTicker.Lock.Unlock()

				feed, err := subscriptionTicker.Parser.ParseURL(subscriptionTicker.URL.String())

				if err != nil {
					subscriptionTicker.Lock.Lock()
					subscriptionTicker.FailedFetches++
					subscriptionTicker.Lock.Unlock()

					log.Warn().Err(err).Msgf("Error in parsing subscription %s, %d failed fetches", subscriptionTicker.URL.String(), subscriptionTicker.FailedFetches)

					if subscriptionTicker.FailedFetches >= 5 {
						for _, sub := range subscriptionTicker.Subscriptions {
							_, _ = readerHandler.Options.BotHandler.Bot.SendMessage(readerHandler.Options.BotHandler.Options.Context, &bot.SendMessageParams{
								ChatID: sub.ChatId,
								Text:   fmt.Sprintf("Could not fetch feed from %s for five times, please check if the fetch source is valid", sub.URL.String()),
							})
						}
					}

					subscriptionTicker.Lock.Lock()

					var httpError gofeed.HTTPError
					ok := errors.As(err, &httpError)
					if ok && httpError.StatusCode == 429 {
						log.Warn().Msgf("Too many requests for %s, retrying later", subscriptionTicker.URL.String())

						timeout := time.Now().Add(readerHandler.Options.WaitTimeout)
						subscriptionTicker.WaitTimeout = &timeout
					}

					subscriptionTicker.InRequest = false
					subscriptionTicker.Lock.Unlock()

					return
				} else {
					log.Trace().Msgf("Handling %s's feed", subscriptionTicker.URL.String())

					_ = readerHandler.handleFeed(subscriptionTicker, feed)
				}

				subscriptionTicker.Lock.Lock()
				subscriptionTicker.InRequest = false
				subscriptionTicker.Lock.Unlock()

				log.Trace().Msgf("Finished fetching %s's feed", subscriptionTicker.URL.String())
			case <-subscriptionTicker.Quit:
				subscriptionTicker.Ticker.Stop()
				return
			}
		}
	}()
}
