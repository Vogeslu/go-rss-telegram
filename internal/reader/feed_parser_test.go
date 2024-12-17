package reader

import (
	"github.com/google/uuid"
	"github.com/mmcdole/gofeed"
	"rss-telegram/internal/subscription"
	"testing"
	"time"
)

func TestFeedParser(t *testing.T) {
	t.Logf("Instantiating mock reader handler")

	readerHandler := NewReaderHandler(&ReaderHandlerOptions{
		RedisDb:             nil,
		BotHandler:          nil,
		SubscriptionHandler: &subscription.SubscriptionHandler{},
	})

	mockItems := getMockFeedItems()

	t.Run("Test shouldSendItem without pattern", func(t *testing.T) {
		sub := &subscription.Subscription{
			Id:            uuid.UUID{},
			ChatId:        0,
			URL:           nil,
			SearchPattern: "",
			CreationDate:  time.Time{},
		}

		matchCount := 0
		expectedCount := len(mockItems)

		for _, item := range mockItems {
			if readerHandler.shouldSendItem(item, sub) {
				matchCount++
			}
		}

		if expectedCount != matchCount {
			t.Errorf("Check mock item count is incorrect, got: %d, want: %d.", matchCount, expectedCount)
		}
	})

	t.Run("Test shouldSendItem with one pattern", func(t *testing.T) {
		sub := &subscription.Subscription{
			Id:            uuid.UUID{},
			ChatId:        0,
			URL:           nil,
			SearchPattern: "News",
			CreationDate:  time.Time{},
		}

		matchCount := 0
		expectedCount := 5

		for _, item := range mockItems {
			if readerHandler.shouldSendItem(item, sub) {
				matchCount++
			}
		}

		if expectedCount != matchCount {
			t.Errorf("Check mock item count is incorrect, got: %d, want: %d.", matchCount, expectedCount)
		}
	})

	t.Run("Test shouldSendItem with multiple pattern", func(t *testing.T) {
		sub := &subscription.Subscription{
			Id:            uuid.UUID{},
			ChatId:        0,
			URL:           nil,
			SearchPattern: "News,Breaking,poll",
			CreationDate:  time.Time{},
		}

		matchCount := 0
		expectedCount := 8

		for _, item := range mockItems {
			if readerHandler.shouldSendItem(item, sub) {
				matchCount++
			}
		}

		if expectedCount != matchCount {
			t.Errorf("Check mock item count is incorrect, got: %d, want: %d.", matchCount, expectedCount)
		}
	})

	t.Run("Test shouldSendItem with multiple pattern (including spaces)", func(t *testing.T) {
		sub := &subscription.Subscription{
			Id:            uuid.UUID{},
			ChatId:        0,
			URL:           nil,
			SearchPattern: "  News,Breaking,  poll",
			CreationDate:  time.Time{},
		}

		matchCount := 0
		expectedCount := 8

		for _, item := range mockItems {
			if readerHandler.shouldSendItem(item, sub) {
				matchCount++
			}
		}

		if expectedCount != matchCount {
			t.Errorf("Check mock item count is incorrect, got: %d, want: %d.", matchCount, expectedCount)
		}
	})
}

func getMockFeedItems() []*gofeed.Item {
	return []*gofeed.Item{
		{Title: "Breaking News Update", Description: "Get the latest breaking news and updates from around the world.", Link: "https://example.com/breaking-news-update"},
		{Title: "Election Poll Results", Description: "A detailed analysis of the recent election poll results and trends.", Link: "https://example.com/election-poll-results"},
		{Title: "Tech Innovations", Description: "Explore the latest in technology and innovations shaping our future.", Link: "https://example.com/tech-innovations"},
		{Title: "Daily News Highlights", Description: "Top news stories and highlights from today's global headlines.", Link: "https://example.com/daily-news-highlights"},
		{Title: "Breaking Insights", Description: "Breaking insights and expert opinions on current events.", Link: "https://example.com/breaking-insights"},
		{Title: "Local Election News", Description: "Coverage of local elections, including breaking updates and insights.", Link: "https://example.com/local-election-news"},
		{Title: "Market Trends", Description: "An in-depth look at market trends and economic developments.", Link: "https://example.com/market-trends"},
		{Title: "Poll Analysis Report", Description: "Comprehensive poll analysis and what it means for the next election.", Link: "https://example.com/poll-analysis-report"},
		{Title: "Entertainment Buzz", Description: "Catch up on the latest entertainment buzz and celebrity news.", Link: "https://example.com/entertainment-buzz"},
		{Title: "Weather Forecast", Description: "Breaking weather news and forecasts for your area and beyond.", Link: "https://example.com/weather-forecast"},
	}
}
