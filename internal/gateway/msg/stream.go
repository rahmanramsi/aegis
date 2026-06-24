package msg

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/go-telegram/bot"
)

const flushInterval = 250 * time.Millisecond

// telegramStream uses Telegram's native sendMessageDraft for smooth animated
// previews (like Hermes), then sendMessage for the final persistent message.
type telegramStream struct {
	adapter *TelegramAdapter
	chatID  int64
	text    string
	mu      sync.Mutex
	dirty   bool
	done    bool
	stopCh  chan struct{}
	draftID string
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
		draftID: fmt.Sprintf("aegis_%d", time.Now().UnixNano()),
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

	// Use native draft — animated preview, no persistent message
	_, err := s.adapter.b.SendMessageDraft(context.Background(), &bot.SendMessageDraftParams{
		ChatID:  s.chatID,
		DraftID: s.draftID,
		Text:    s.text,
	})
	if err != nil {
		slog.Warn("stream: draft", "err", err)
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
	time.Sleep(flushInterval + 50*time.Millisecond)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Send final persistent message (clears the draft automatically)
	if s.text != "" {
		_, err := s.adapter.b.SendMessage(context.Background(), &bot.SendMessageParams{
			ChatID: s.chatID,
			Text:   s.text,
		})
		if err != nil {
			slog.Warn("stream: final send", "err", err)
		}
	} else {
		s.adapter.b.SendMessage(context.Background(), &bot.SendMessageParams{
			ChatID: s.chatID,
			Text:   "Done",
		})
	}
	return nil
}

func (s *telegramStream) Error(text string) error {
	close(s.stopCh)
	time.Sleep(flushInterval + 50*time.Millisecond)

	s.adapter.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: s.chatID,
		Text:   text,
	})
	return nil
}

var _ StreamSender = (*telegramStream)(nil)
