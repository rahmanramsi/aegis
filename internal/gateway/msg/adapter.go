package msg

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
	Close() error
}
