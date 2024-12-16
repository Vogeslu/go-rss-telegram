package utils

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"strings"
)

func SendChunkedMessage(text string, ctx context.Context, b *bot.Bot, chatId int64, chunkSize int, replyMarkup models.ReplyMarkup) {
	lines := strings.Split(text, "\n")
	var chunk string

	appendReplyMarkup := replyMarkup

	for _, line := range lines {
		for len(line) > chunkSize {
			part := line[:chunkSize]
			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatId,
				Text:        part,
				ReplyMarkup: appendReplyMarkup,
			})
			line = line[chunkSize:]
			appendReplyMarkup = nil
		}

		if len(chunk)+len(line)+1 > chunkSize {
			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatId,
				Text:        chunk,
				ReplyMarkup: appendReplyMarkup,
			})
			chunk = ""
			appendReplyMarkup = nil
		}

		if chunk != "" {
			chunk += "\n"
		}
		chunk += line
	}

	if chunk != "" {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatId,
			Text:        chunk,
			ReplyMarkup: appendReplyMarkup,
		})
	}
}
