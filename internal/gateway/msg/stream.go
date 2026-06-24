package msg

import (
	"context"
	"log/slog"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-telegram/bot"
)

const (
	flushInterval  = 300 * time.Millisecond
	bufferMinChars = 8
	maxFloodErrors = 3
	cursorChar     = " ▍"
)

var draftIDCounter atomic.Int64

type telegramStream struct {
	adapter *TelegramAdapter
	chatID  int64

	text   string
	mu     sync.Mutex
	dirty  bool
	stopCh chan struct{}

	draftID       int64
	draftsEnabled bool
	floodErrors   int

	msgID     int
	sentFirst bool
}

func (t *TelegramAdapter) SendStream(chatID string) StreamSender {
	chatIDInt, _ := strconv.ParseInt(chatID, 10, 64)
	draftIDCounter.Add(1)
	s := &telegramStream{
		adapter:       t,
		chatID:        chatIDInt,
		stopCh:        make(chan struct{}),
		draftID:       draftIDCounter.Load(),
		draftsEnabled: true,
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

	if !s.dirty || len(s.text) < bufferMinChars {
		return
	}
	s.dirty = false

	// Hermes: apply MarkdownV2 + cursor
	formatted := FormatTelegramMarkdown(s.text) + cursorChar

	if s.draftsEnabled {
		s.flushDraft(formatted)
	} else {
		s.flushEdit(formatted)
	}
}

func (s *telegramStream) flushDraft(text string) {
	_, err := s.adapter.b.SendMessageDraft(context.Background(), &bot.SendMessageDraftParams{
		ChatID:  s.chatID,
		DraftID: strconv.FormatInt(s.draftID, 10),
		Text:    text,
	})
	if err != nil {
		s.floodErrors++
		if s.floodErrors >= maxFloodErrors {
			s.draftsEnabled = false
			slog.Warn("stream: drafts disabled, falling back to edit path")
			s.sendInitialLocked(FormatTelegramMarkdown(s.text) + cursorChar)
		}
		return
	}
	s.floodErrors = 0
}

func (s *telegramStream) flushEdit(text string) {
	if !s.sentFirst {
		s.sendInitialLocked(text)
		return
	}
	_, err := s.adapter.b.EditMessageText(context.Background(), &bot.EditMessageTextParams{
		ChatID:    s.chatID,
		MessageID: s.msgID,
		Text:      text,
	})
	if err != nil {
		slog.Warn("stream: edit failed", "err", err)
	}
}

func (s *telegramStream) sendInitialLocked(text string) {
	msg, err := s.adapter.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: s.chatID,
		Text:   text,
	})
	if err != nil {
		slog.Warn("stream: send failed", "err", err)
		return
	}
	s.msgID = msg.ID
	s.sentFirst = true
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
	time.Sleep(flushInterval + 200*time.Millisecond)

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.text == "" {
		s.adapter.b.SendMessage(context.Background(), &bot.SendMessageParams{
			ChatID: s.chatID,
			Text:   "Done",
		})
		return nil
	}

	final := FormatTelegramMarkdown(s.text)

	if s.draftsEnabled {
		s.adapter.b.SendMessage(context.Background(), &bot.SendMessageParams{
			ChatID: s.chatID,
			Text:   final,
		})
	} else {
		_, err := s.adapter.b.EditMessageText(context.Background(), &bot.EditMessageTextParams{
			ChatID:    s.chatID,
			MessageID: s.msgID,
			Text:      final,
		})
		if err != nil {
			s.adapter.b.SendMessage(context.Background(), &bot.SendMessageParams{
				ChatID: s.chatID,
				Text:   final,
			})
		}
	}
	return nil
}

func (s *telegramStream) Error(text string) error {
	close(s.stopCh)
	time.Sleep(flushInterval + 200*time.Millisecond)

	s.adapter.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: s.chatID,
		Text:   text,
	})
	return nil
}

var _ StreamSender = (*telegramStream)(nil)
