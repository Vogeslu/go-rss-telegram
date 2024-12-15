# go-rss-telegram

Telegram bot with rss-feed-subscriptions written in Go

## Installation

### Prerequisites

- [Redis](https://redis.io/)
- [Go 1.23](https://go.dev/)
- [Telegram Bot](https://t.me/BotFather)

### Setup

```bash
# 1. Pull repository
$ git pull https://github.com/Vogeslu/go-rss-telegram

# 2. Open repository
$ cd go-rss-telegram

# 3. Create rss-telegram.env for config (see sample)
$ touch rss-telegram.env

# 4. Start service
$ go run ./cmd/rss-telegram/main.go

# 5. Start conversation with bot in telegram
```

### Configuration sample

```env
BOT_TOKEN=YOUR_TOKEN
LOG_LEVEL=debug
RSS_INTERVAL=5 # Check for new items every 5 seconds
```

## Available commands via telegram

- `/start` - Initial command
- `/subscribe` - Subscribe to a new feed
- `/unsubscribe` - Unsubscribe from feed
- `/subscriptions` - List subscriptions