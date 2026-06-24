package msg

import (
	"context"
	"log/slog"
	"strconv"
	"sync"

	"github.com/go-telegram/bot"
)
// telegramStream accumulates stdout chunks and edits a single Telegram message.
type telegramStream struct {
	adapter   *TelegramAdapter
	chatID    int64
	messageID int
	text      string
	mu        sync.Mutex
	sentFirst bool
}

func (t *TelegramAdapter) SendStream(chatID string) StreamSender {
	chatIDInt, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		slog.Warn("stream: invalid chat id", "chat_id", chatID)
		chatIDInt = 0
	}
	return &telegramStream{adapter: t, chatID: chatIDInt}
}

func (s *telegramStream) Append(text string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.text += text

	if !s.sentFirst {
		msg, err := s.adapter.b.SendMessage(context.Background(), &bot.SendMessageParams{
			ChatID: s.chatID,
			Text:   s.text,
		})
		if err != nil {
			slog.Warn("stream: send initial", "err", err)
			return nil
		}
		s.messageID = msg.ID
		s.sentFirst = true
		return nil
	}

	_, err := s.adapter.b.EditMessageText(context.Background(), &bot.EditMessageTextParams{
		ChatID:    s.chatID,
		MessageID: s.messageID,
		Text:      s.text,
	})
	if err != nil {
		slog.Warn("stream: edit", "err", err)
	}
	return nil
}

func (s *telegramStream) Done() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.sentFirst && s.text == "" {
		s.adapter.b.SendMessage(context.Background(), &bot.SendMessageParams{
			ChatID: s.chatID,
			Text:   "✓ Done",
		})
		return nil
	}

	// Final edit to ensure complete text is shown
	if s.messageID != 0 {
		_, err := s.adapter.b.EditMessageText(context.Background(), &bot.EditMessageTextParams{
			ChatID:    s.chatID,
			MessageID: s.messageID,
			Text:      s.text,
		})
		if err != nil {
			slog.Warn("stream: final edit", "err", err)
			// If edit fails (message deleted, etc), send as new message
			s.adapter.b.SendMessage(context.Background(), &bot.SendMessageParams{
				ChatID: s.chatID,
				Text:   s.text,
			})
		}
	}
	return nil
}

func (s *telegramStream) Error(text string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.messageID != 0 {
		edited := s.text + "\n\n❌ " + text
		s.adapter.b.EditMessageText(context.Background(), &bot.EditMessageTextParams{
			ChatID:    s.chatID,
			MessageID: s.messageID,
			Text:      edited,
		})
	} else {
		s.adapter.b.SendMessage(context.Background(), &bot.SendMessageParams{
			ChatID: s.chatID,
			Text:   "❌ " + text,
		})
	}
	return nil
}

var _ StreamSender = (*telegramStream)(nil)
