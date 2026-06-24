package msg

import (
	"context"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/go-telegram/bot"
)

const flushInterval = 200 * time.Millisecond

// telegramStream buffers text and edits a single Telegram message at intervals,
// creating a smooth streaming typing effect.
type telegramStream struct {
	adapter    *TelegramAdapter
	chatID     int64
	messageID  int
	text       string
	mu         sync.Mutex
	sentFirst  bool
	dirty      bool
	done       bool
	flushCh    chan struct{}
	stopCh     chan struct{}
	flushTimer *time.Timer
}

func (t *TelegramAdapter) SendStream(chatID string) StreamSender {
	chatIDInt, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		slog.Warn("stream: invalid chat id", "chat_id", chatID)
		chatIDInt = 0
	}
	s := &telegramStream{
		adapter: t,
		chatID:  chatIDInt,
		stopCh:  make(chan struct{}),
	}
	go s.flushLoop()
	return s
}

func (s *telegramStream) flushLoop() {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.flush()
		case <-s.stopCh:
			s.flush()
			return
		}
	}
}

func (s *telegramStream) flush() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.dirty || s.text == "" {
		return
	}
	s.dirty = false

	if !s.sentFirst {
		msg, err := s.adapter.b.SendMessage(context.Background(), &bot.SendMessageParams{
			ChatID: s.chatID,
			Text:   s.text,
			LinkPreviewOptions: &bot.LinkPreviewOptions{IsDisabled: true},
		})
		if err != nil {
			slog.Warn("stream: send initial", "err", err)
			return
		}
		s.messageID = msg.ID
		s.sentFirst = true
		return
	}

	_, err := s.adapter.b.EditMessageText(context.Background(), &bot.EditMessageTextParams{
		ChatID:      s.chatID,
		MessageID:   s.messageID,
		Text:        s.text,
		LinkPreview: &bot.LinkPreviewOptions{IsDisabled: true},
	})
	if err != nil {
		slog.Warn("stream: edit", "err", err)
	}
}

func (s *telegramStream) Append(text string) error {
	s.mu.Lock()
	s.text += text
	s.dirty = true
	s.mu.Unlock()
	return nil
}

func (s *telegramStream) Done() error {
	close(s.stopCh)
	// Wait for final flush
	time.Sleep(flushInterval + 50*time.Millisecond)

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.sentFirst && s.text == "" {
		s.adapter.b.SendMessage(context.Background(), &bot.SendMessageParams{
			ChatID: s.chatID,
			Text:   "✓ Done",
		})
		return nil
	}

	// Final edit with complete text
	if s.messageID != 0 {
		_, err := s.adapter.b.EditMessageText(context.Background(), &bot.EditMessageTextParams{
			ChatID:    s.chatID,
			MessageID: s.messageID,
			Text:      s.text,
		})
		if err != nil {
			s.adapter.b.SendMessage(context.Background(), &bot.SendMessageParams{
				ChatID: s.chatID,
				Text:   s.text,
			})
		}
	}
	return nil
}

func (s *telegramStream) Error(text string) error {
	close(s.stopCh)
	time.Sleep(flushInterval + 50*time.Millisecond)

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.messageID != 0 {
		edited := s.text + "\n\n\u274c " + text
		s.adapter.b.EditMessageText(context.Background(), &bot.EditMessageTextParams{
			ChatID:    s.chatID,
			MessageID: s.messageID,
			Text:      edited,
		})
	} else {
		s.adapter.b.SendMessage(context.Background(), &bot.SendMessageParams{
			ChatID: s.chatID,
			Text:   "\u274c " + text,
		})
	}
	return nil
}

var _ StreamSender = (*telegramStream)(nil)
