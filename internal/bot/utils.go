package bot

import (
	"context"
	"github.com/go-telegram/bot"
)

func sendMessage(b *bot.Bot, ctx context.Context, chatId int64, message string) error {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatId,
		Text:   message,
	})

	return err
}
