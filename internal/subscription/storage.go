package subscription

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/url"
	"time"
)

type Subscription struct {
	Id            uuid.UUID `json:"id"`
	ChatId        int64     `json:"chatId"`
	URL           *url.URL  `json:"url"`
	SearchPattern string    `json:"searchPattern"`
	CreationDate  time.Time `json:"creationDate"`
}

func (subscriptionHandler *SubscriptionHandler) AddSubscription(chatId int64, subscription *Subscription) (string, error) {
	key := fmt.Sprintf("subscription:%d:%s", chatId, subscription.Id.String())

	subscriptionBytes, err := json.Marshal(subscription)
	if err != nil {
		return "", err
	}

	err = subscriptionHandler.Options.RedisDb.Set(subscriptionHandler.Context, key, subscriptionBytes, 0).Err()
	if err != nil {
		return "", err
	}

	if subscriptionHandler.ReaderEventListener != nil {
		subscriptionHandler.ReaderEventListener.AddSubscription(subscription)
	}

	return key, err
}

func (subscriptionHandler *SubscriptionHandler) DeleteSubscription(chatId int64, subscription *Subscription) {
	_ = subscriptionHandler.Options.RedisDb.Del(subscriptionHandler.Context, fmt.Sprintf("subscription:%d:%s", chatId, subscription.Id.String())).Err()
	_ = subscriptionHandler.Options.RedisDb.Del(subscriptionHandler.Context, fmt.Sprintf("post-fetch:%d:%s", chatId, subscription.Id.String())).Err()
	_ = subscriptionHandler.Options.RedisDb.Del(subscriptionHandler.Context, fmt.Sprintf("guids:%d:%s", chatId, subscription.Id.String())).Err()

	if subscriptionHandler.ReaderEventListener != nil {
		subscriptionHandler.ReaderEventListener.RemoveSubscription(subscription)
	}
}

func (subscriptionHandler *SubscriptionHandler) NewSubscription(url *url.URL, chatId int64, searchPattern string) *Subscription {
	return &Subscription{
		Id:            uuid.New(),
		ChatId:        chatId,
		URL:           url,
		SearchPattern: searchPattern,
		CreationDate:  time.Now(),
	}
}

func (subscriptionHandler *SubscriptionHandler) GetAllSubscriptions() ([]*Subscription, error) {
	return subscriptionHandler.GetSubscriptions("*")
}

func (subscriptionHandler *SubscriptionHandler) GetSubscriptionsFromChat(chatId int64) ([]*Subscription, error) {
	return subscriptionHandler.GetSubscriptions(fmt.Sprintf("%d:*", chatId))
}

func (subscriptionHandler *SubscriptionHandler) GetSubscriptions(suffix string) ([]*Subscription, error) {
	keys, err := subscriptionHandler.Options.RedisDb.Keys(subscriptionHandler.Context, fmt.Sprintf("subscription:%s", suffix)).Result()
	if err != nil {
		return nil, err
	}

	output := make([]*Subscription, len(keys))
	for i, key := range keys {
		foundItem, ok := subscriptionHandler.subscriptionsCache[key]
		if ok {
			output[i] = foundItem
			continue
		}

		val, err := subscriptionHandler.Options.RedisDb.Get(subscriptionHandler.Context, key).Result()
		if err != nil {
			return nil, err
		}

		var subscription Subscription
		err = json.Unmarshal([]byte(val), &subscription)
		if err != nil {
			return nil, err
		}

		subscriptionHandler.lock.Lock()
		subscriptionHandler.subscriptionsCache[key] = &subscription
		subscriptionHandler.lock.Unlock()

		output[i] = &subscription
	}

	return output, nil
}

func (subscription *Subscription) String() string {
	urlString := subscription.URL.String()
	date := subscription.CreationDate.Format("01-02-2006 15:04:05")

	patternText := "without pattern"
	if subscription.SearchPattern != "" {
		patternText = fmt.Sprintf("with pattern %s", subscription.SearchPattern)
	}

	return fmt.Sprintf("%s %s, added %s", urlString, patternText, date)
}
