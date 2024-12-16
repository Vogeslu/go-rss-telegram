package reader

import (
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog/log"
	"rss-telegram/internal/subscription"
	"rss-telegram/internal/utils"
	"slices"
	"strings"
)

func (readerHandler *ReaderHandler) handleFeed(subscriptionTicker *SubscriptionTicker, feed *gofeed.Feed) error {
	for _, sub := range subscriptionTicker.Subscriptions {
		newItems, err := readerHandler.getNewItems(feed, sub)
		if err != nil {
			return err
		}

		err = readerHandler.addGuids(newItems, sub)
		if err != nil {
			return err
		}

		firstFetch, err := readerHandler.isFirstFetch(sub)
		if err != nil {
			return err
		}

		if !firstFetch {
			readerHandler.notifyNewItems(newItems, sub)
		}
	}

	return nil
}

func (readerHandler *ReaderHandler) getNewItems(feed *gofeed.Feed, subscription *subscription.Subscription) ([]*gofeed.Item, error) {
	knownGuids, err := readerHandler.Options.RedisDb.SMembers(readerHandler.Context, fmt.Sprintf("guids:%d:%s", subscription.ChatId, subscription.Id)).Result()
	if err != nil {
		return nil, err
	}

	var output []*gofeed.Item
	for _, item := range feed.Items {
		if slices.Contains(knownGuids, item.GUID) {
			continue
		}

		output = append(output, item)
	}

	return output, nil
}

func (readerHandler *ReaderHandler) addGuids(items []*gofeed.Item, subscription *subscription.Subscription) error {
	guids := make([]string, len(items))
	for i, item := range items {
		guids[i] = item.GUID
	}

	err := readerHandler.Options.RedisDb.SAdd(readerHandler.Context, fmt.Sprintf("guids:%d:%s", subscription.ChatId, subscription.Id), guids).Err()
	if err != nil {
		return err
	}

	return nil
}

func (readerHandler *ReaderHandler) notifyNewItems(items []*gofeed.Item, subscription *subscription.Subscription) {
	for _, item := range items {
		if !readerHandler.shouldSendItem(item, subscription) {
			continue
		}

		log.Trace().Msg(itemAsMessage(item))

		_, _ = readerHandler.Options.BotHandler.Bot.SendMessage(readerHandler.Options.BotHandler.Options.Context, &bot.SendMessageParams{
			ChatID:    subscription.ChatId,
			Text:      itemAsMessage(item),
			ParseMode: models.ParseModeHTML,
		})

	}
}

func (readerHandler *ReaderHandler) shouldSendItem(item *gofeed.Item, subscription *subscription.Subscription) bool {
	pattern := subscription.SearchPattern

	if pattern == "" {
		return true
	}

	for _, patternItem := range strings.Split(pattern, ",") {
		patternItem = strings.Trim(patternItem, " ")

		if utils.ContainsInsensitive(item.Title, patternItem) {
			return true
		}

		if utils.ContainsInsensitive(item.Description, patternItem) {
			return true
		}

		if utils.ContainsInsensitive(item.Link, patternItem) {
			return true
		}
	}

	return false
}

func (readerHandler *ReaderHandler) isFirstFetch(subscription *subscription.Subscription) (bool, error) {
	val, err := readerHandler.Options.RedisDb.Exists(readerHandler.Context, fmt.Sprintf("post-fetch:%d:%s", subscription.ChatId, subscription.Id)).Result()
	if err != nil {
		return false, err
	}

	if val == 0 {
		err = readerHandler.Options.RedisDb.Set(readerHandler.Context, fmt.Sprintf("post-fetch:%d:%s", subscription.ChatId, subscription.Id), "1", 0).Err()
		if err != nil {
			return false, err
		}
	}

	return val == 0, nil
}

func itemAsMessage(item *gofeed.Item) string {
	var output []string

	if item.Title != "" {
		output = append(output, fmt.Sprintf("<b>%s</b>\n", item.Title))
	}

	if item.Description != "" {
		output = append(output, fmt.Sprintf("%s", item.Description))
	}

	if item.Link != "" {
		output = append(output, fmt.Sprintf("\n%s", item.Link))
	}

	return strings.Join(output, "\n")
}
