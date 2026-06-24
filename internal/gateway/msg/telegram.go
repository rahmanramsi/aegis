package msg

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type TelegramAdapter struct {
	token string
	b     *bot.Bot
	msgCh chan Message
}

func NewTelegramAdapter() (*TelegramAdapter, error) {
	token := os.Getenv("AEGIS_TELEGRAM_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("AEGIS_TELEGRAM_TOKEN not set")
	}

	b, err := bot.New(token)
	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}

	return &TelegramAdapter{
		token: token,
		b:     b,
		msgCh: make(chan Message, 64),
	}, nil
}

func (t *TelegramAdapter) Start(ctx context.Context) (<-chan Message, error) {
	go func() {
		<-ctx.Done()
		t.b.Close(ctx)
	}()

	t.b.RegisterHandlerMatchFunc(
		func(update *models.Update) bool {
			return update.Message != nil && update.Message.Text != ""
		},
		func(ctx context.Context, b *bot.Bot, update *models.Update) {
			chatID := strconv.FormatInt(update.Message.Chat.ID, 10)
			userID := strconv.FormatInt(update.Message.From.ID, 10)
			userName := update.Message.From.FirstName

			t.msgCh <- Message{
				Platform: "telegram",
				ChatID:   chatID,
				UserID:   userID,
				UserName: userName,
				Text:     update.Message.Text,
			}
		},
	)

	// Start long polling in background (blocking call)
	go t.b.Start(ctx)

	slog.Info("telegram adapter started")
	return t.msgCh, nil
}

func (t *TelegramAdapter) Send(chatID, text string) error {
	chatIDInt, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		slog.Warn("telegram send: invalid chat id", "chat_id", chatID, "err", err)
		return nil
	}
	slog.Info("telegram send", "chat_id", chatID, "text", text)
	_, err = t.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: chatIDInt,
		Text:   text,
	})
	if err != nil {
		slog.Warn("telegram send failed", "chat_id", chatID, "err", err)
	}
	return nil
}

func (t *TelegramAdapter) Close() error {
	return nil
}
