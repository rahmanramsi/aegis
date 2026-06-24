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
	bufferMinChars = 8  // Hermes: don't flush until enough content
	maxFloodErrors = 3  // Hermes: consecutive failures → fallback to edit path
	cursorChar     = " ▌"
)

// Hermes pattern: class-level monotonic counter. Reusing the same draft_id
// across consecutive calls triggers Telegram's native draft animation.
var draftIDCounter atomic.Int64

type telegramStream struct {
	adapter *TelegramAdapter
	chatID  int64

	// buffered text (accumulated from Append calls)
	text string
	mu   sync.Mutex

	// dirty flag — set by Append, cleared by flush
	dirty bool

	// lifecycle
	stopCh chan struct{}

	// draft streaming state
	draftID       int64
	draftsEnabled bool
	floodErrors   int

	// legacy edit fallback state (when drafts disabled)
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
			s.flush() // final flush before exit
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

	// Hermes pattern: append blinking cursor during streaming, strip on final
	displayText := s.text + cursorChar

	if s.draftsEnabled {
		s.flushDraft(displayText)
	} else {
		s.flushEdit(displayText)
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
		slog.Warn("stream: draft failed", "err", err, "flood", s.floodErrors)
		if s.floodErrors >= maxFloodErrors {
			s.draftsEnabled = false
			slog.Warn("stream: drafts disabled — falling back to edit path")
			// Send current text as initial edit-path message
			s.sendInitialLocked(s.text + cursorChar)
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
		slog.Warn("stream: initial send failed", "err", err)
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

// Done sends the final persistent message and stops the flush loop.
// If drafts were used, the draft auto-clears when sendMessage arrives.
func (s *telegramStream) Done() error {
	close(s.stopCh)
	// Give the flush loop time to drain
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

	if s.draftsEnabled {
		// Draft path: send final persistent message (draft auto-clears)
		s.adapter.b.SendMessage(context.Background(), &bot.SendMessageParams{
			ChatID: s.chatID,
			Text:   s.text,
		})
	} else {
		// Edit path: final edit without cursor, or send new if edit fails
		_, err := s.adapter.b.EditMessageText(context.Background(), &bot.EditMessageTextParams{
			ChatID:    s.chatID,
			MessageID: s.msgID,
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
	time.Sleep(flushInterval + 200*time.Millisecond)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.adapter.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: s.chatID,
		Text:   text,
	})
	return nil
}

var _ StreamSender = (*telegramStream)(nil)
