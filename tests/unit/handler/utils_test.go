package handler_test

import (
	"testing"

	"github.com/go-telegram/bot/models"
)

func callbackChatMessage(upd *models.Update) (int64, int, bool) {
	if upd == nil || upd.CallbackQuery == nil {
		return 0, 0, false
	}
	m := upd.CallbackQuery.Message
	if m.Message != nil {
		return m.Message.Chat.ID, m.Message.ID, true
	}
	if m.InaccessibleMessage != nil {
		return m.InaccessibleMessage.Chat.ID, m.InaccessibleMessage.MessageID, true
	}
	return 0, 0, false
}

func TestCallbackChatMessage(t *testing.T) {
	upd := &models.Update{
		CallbackQuery: &models.CallbackQuery{
			Message: models.MaybeInaccessibleMessage{
				Message: &models.Message{ID: 1, Chat: models.Chat{ID: 42}},
			},
		},
	}
	chatID, msgID, ok := callbackChatMessage(upd)
	if !ok || chatID != 42 || msgID != 1 {
		t.Fatalf("expected (42,1,true) got (%d,%d,%v)", chatID, msgID, ok)
	}

	upd.CallbackQuery.Message = models.MaybeInaccessibleMessage{
		InaccessibleMessage: &models.InaccessibleMessage{Chat: models.Chat{ID: 7}, MessageID: 8},
	}
	chatID, msgID, ok = callbackChatMessage(upd)
	if !ok || chatID != 7 || msgID != 8 {
		t.Fatalf("expected (7,8,true) got (%d,%d,%v)", chatID, msgID, ok)
	}

	upd.CallbackQuery = nil
	if _, _, ok := callbackChatMessage(upd); ok {
		t.Fatal("expected false for nil callback")
	}
}
