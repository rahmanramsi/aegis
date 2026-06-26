package messaging

import (
	"context"
	"log/slog"
	"sync"
)

type TokenRouter func(ctx context.Context, tokenHash string, m Message, adapter Adapter)

type BotManager struct {
	mu       sync.Mutex
	adapters map[string]*botInstance // tokenHash → adapter
	routeFn  TokenRouter
}

type botInstance struct {
	adapter *TelegramAdapter
	cancel  context.CancelFunc
}

func NewBotManager(routeFn TokenRouter) *BotManager {
	return &BotManager{
		adapters: make(map[string]*botInstance),
		routeFn:  routeFn,
	}
}

func (bm *BotManager) AddBot(ctx context.Context, token string) error {
	hash := sha256Hex(token)

	bm.mu.Lock()
	if _, exists := bm.adapters[hash]; exists {
		bm.mu.Unlock()
		slog.Warn("bot already running", "token_hash", hash[:8])
		return nil
	}

	adapter, err := NewTelegramAdapterWithToken(token)
	if err != nil {
		bm.mu.Unlock()
		return err
	}

	botCtx, cancel := context.WithCancel(ctx)
	bi := &botInstance{adapter: adapter, cancel: cancel}
	bm.adapters[hash] = bi
	bm.mu.Unlock()

	msgCh, err := adapter.Start(botCtx)
	if err != nil {
		bm.RemoveBot(token)
		return err
	}

	tokenHash := hash
	adpt := adapter
	go func() {
		for m := range msgCh {
			bm.routeFn(botCtx, tokenHash, m, adpt)
		}
	}()

	slog.Info("telegram bot started", "hash", hash[:8])
	return nil
}

func (bm *BotManager) RemoveBot(token string) {
	hash := sha256Hex(token)
	bm.mu.Lock()
	bi, ok := bm.adapters[hash]
	if ok {
		delete(bm.adapters, hash)
	}
	bm.mu.Unlock()

	if ok && bi.cancel != nil {
		bi.cancel()
		bi.adapter.Close()
		slog.Info("telegram bot stopped", "hash", hash[:8])
	}
}

func (bm *BotManager) Send(chatID, text string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	for _, bi := range bm.adapters {
		return bi.adapter.Send(chatID, text)
	}
	return nil
}

func (bm *BotManager) RemoveBotByHash(hash string) {
	bm.mu.Lock()
	bi, ok := bm.adapters[hash]
	if ok {
		delete(bm.adapters, hash)
	}
	bm.mu.Unlock()
	if ok && bi.cancel != nil {
		bi.cancel()
		bi.adapter.Close()
		slog.Info("telegram bot stopped", "hash", hash[:8])
	}
}
