package messaging

import "context"

type Message struct {
	Platform string
	ChatID   string
	UserID   string
	UserName string
	Text     string
}

type Adapter interface {
	Start(ctx context.Context) (<-chan Message, error)
	Send(chatID, text string) error
	SendTyping(chatID string) error
	SendStream(chatID string) StreamSender
	Close() error
}

// StreamSender accumulates text and edits a single Telegram message.
type StreamSender interface {
	Append(text string) error
	Done() error
	Error(text string) error
}
