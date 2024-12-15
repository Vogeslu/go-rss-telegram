package reader

import (
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog/log"
	"rss-telegram/internal/utils"
	"slices"
	"strings"
)

func (readerHandler *ReaderHandler) handleFeed(subscriptionTicker *SubscriptionTicker, feed *gofeed.Feed) error {
	newItems, err := readerHandler.getNewItems(subscriptionTicker, feed)
	if err != nil {
		return err
	}

	err = readerHandler.addGuids(subscriptionTicker, newItems)
	if err != nil {
		return err
	}

	firstFetch, err := readerHandler.isFirstFetch(subscriptionTicker)
	if err != nil {
		return err
	}

	if !firstFetch {
		readerHandler.notifyNewItems(subscriptionTicker, newItems)
	}

	return nil
}

func (readerHandler *ReaderHandler) getNewItems(subscriptionTicker *SubscriptionTicker, feed *gofeed.Feed) ([]*gofeed.Item, error) {
	knownGuids, err := readerHandler.Options.RedisDb.SMembers(readerHandler.Context, fmt.Sprintf("guids:%d:%s", subscriptionTicker.Subscription.ChatId, subscriptionTicker.Subscription.Id)).Result()
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

func (readerHandler *ReaderHandler) addGuids(subscriptionTicker *SubscriptionTicker, items []*gofeed.Item) error {
	guids := make([]string, len(items))
	for i, item := range items {
		guids[i] = item.GUID
	}

	err := readerHandler.Options.RedisDb.SAdd(readerHandler.Context, fmt.Sprintf("guids:%d:%s", subscriptionTicker.Subscription.ChatId, subscriptionTicker.Subscription.Id), guids).Err()
	if err != nil {
		return err
	}

	return nil
}

func (readerHandler *ReaderHandler) notifyNewItems(subscriptionTicker *SubscriptionTicker, items []*gofeed.Item) {
	for _, item := range items {
		if !readerHandler.shouldSendItem(subscriptionTicker, item) {
			continue
		}

		log.Trace().Msg(itemAsMessage(item))

		_, _ = readerHandler.Options.BotHandler.Bot.SendMessage(readerHandler.Options.BotHandler.Options.Context, &bot.SendMessageParams{
			ChatID:    subscriptionTicker.Subscription.ChatId,
			Text:      itemAsMessage(item),
			ParseMode: models.ParseModeHTML,
		})

	}
}

func (readerHandler *ReaderHandler) shouldSendItem(subscriptionTicker *SubscriptionTicker, item *gofeed.Item) bool {
	pattern := subscriptionTicker.Subscription.SearchPattern

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

func (readerHandler *ReaderHandler) isFirstFetch(subscriptionTicker *SubscriptionTicker) (bool, error) {
	val, err := readerHandler.Options.RedisDb.Exists(readerHandler.Context, fmt.Sprintf("post-fetch:%d:%s", subscriptionTicker.Subscription.ChatId, subscriptionTicker.Subscription.Id)).Result()
	if err != nil {
		return false, err
	}

	if val == 0 {
		err = readerHandler.Options.RedisDb.Set(readerHandler.Context, fmt.Sprintf("post-fetch:%d:%s", subscriptionTicker.Subscription.ChatId, subscriptionTicker.Subscription.Id), "1", 0).Err()
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
