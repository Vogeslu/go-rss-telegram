package main

import (
	"context"
	"github.com/mxcd/go-config/config"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"rss-telegram/internal/bot"
	"rss-telegram/internal/chats"
	"rss-telegram/internal/reader"
	"rss-telegram/internal/subscription"
	"rss-telegram/internal/utils"
	"time"
)

func main() {
	err := utils.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed loading config")
	}

	config.Print()

	logLevel, err := zerolog.ParseLevel(config.Get().String("LOG_LEVEL"))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed parsing log level")
	}

	zerolog.SetGlobalLevel(logLevel)

	log.Info().Msg("Starting redis client...")

	redisDb, err := setupRedis()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed pinging redis server")
	}

	subscriptionHandler := subscription.NewSubscriptionHandler(&subscription.SubscriptionHandlerOptions{
		RedisDb: redisDb,
	})

	chatHandler := chats.NewChatHandler(&chats.ChatHandlerOptions{
		RedisDb:             redisDb,
		SubscriptionHandler: subscriptionHandler,
	})

	log.Info().Msg("Starting bot...")

	botCtx, botCancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer botCancel()

	botHandler, err := bot.NewBotHandler(&bot.BotHandlerOptions{
		RedisDb:             redisDb,
		BotToken:            config.Get().String("BOT_TOKEN"),
		ChatHandler:         chatHandler,
		SubscriptionHandler: subscriptionHandler,
		Context:             botCtx,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed starting bot")
	}

	readerHandler := reader.NewReaderHandler(&reader.ReaderHandlerOptions{
		RedisDb:             redisDb,
		BotHandler:          botHandler,
		SubscriptionHandler: subscriptionHandler,
		Interval:            time.Duration(config.Get().Int("RSS_INTERVAL")) * time.Second,
		WaitTimeout:         time.Duration(config.Get().Int("RSS_429_TIMEOUT")) * time.Second,
	})

	err = readerHandler.AddSubscriptions()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed adding subscriptions to reader")
	}

	readerHandler.RunSubscriptions()

	botHandler.Bot.Start(botCtx)
}

func setupRedis() (*redis.Client, error) {
	redisDb := redis.NewClient(&redis.Options{
		Addr:     config.Get().String("REDIS_HOST"),
		Password: config.Get().String("REDIS_PASSWORD"),
		DB:       config.Get().Int("REDIS_DB"),
	})

	err := redisDb.Ping(context.Background()).Err()
	if err != nil {
		return nil, err
	}

	return redisDb, nil
}
