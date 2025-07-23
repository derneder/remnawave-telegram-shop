package app

import (
	"context"
	"testing"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"remnawave-tg-shop-bot/internal/adapter/telegram/handler"
	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	"remnawave-tg-shop-bot/internal/pkg/translation"
)

type fakeCustomerRepo struct{ calls int }

func (f *fakeCustomerRepo) FindById(ctx context.Context, id int64) (*domaincustomer.Customer, error) {
	return nil, nil
}
func (f *fakeCustomerRepo) FindByTelegramId(ctx context.Context, id int64) (*domaincustomer.Customer, error) {
	f.calls++
	return &domaincustomer.Customer{TelegramID: id}, nil
}
func (f *fakeCustomerRepo) Create(ctx context.Context, c *domaincustomer.Customer) (*domaincustomer.Customer, error) {
	return c, nil
}
func (f *fakeCustomerRepo) UpdateFields(ctx context.Context, id int64, updates map[string]interface{}) error {
	return nil
}
func (f *fakeCustomerRepo) FindByTelegramIds(ctx context.Context, ids []int64) ([]domaincustomer.Customer, error) {
	return nil, nil
}
func (f *fakeCustomerRepo) DeleteByNotInTelegramIds(ctx context.Context, ids []int64) error {
	return nil
}
func (f *fakeCustomerRepo) CreateBatch(ctx context.Context, customers []domaincustomer.Customer) error {
	return nil
}
func (f *fakeCustomerRepo) UpdateBatch(ctx context.Context, customers []domaincustomer.Customer) error {
	return nil
}
func (f *fakeCustomerRepo) FindByExpirationRange(ctx context.Context, startDate, endDate time.Time) (*[]domaincustomer.Customer, error) {
	return nil, nil
}

func TestInitHandlers(t *testing.T) {
	b, err := bot.New("1:1", bot.WithSkipGetMe(), bot.WithNotAsyncHandlers())
	if err != nil {
		t.Fatal(err)
	}

	tm := translation.GetInstance()
	if err := tm.InitDefaultTranslations(); err != nil {
		t.Fatal(err)
	}

	repo := &fakeCustomerRepo{}
	h := handler.NewHandler(nil, nil, tm, repo, nil, nil, nil, nil, nil, nil, nil)

	initHandlers(b, h)

	upd := &models.Update{
		ID: 1,
		Message: &models.Message{
			Chat:     models.Chat{ID: 1},
			From:     &models.User{ID: 1, LanguageCode: "en"},
			Text:     "/connect",
			Entities: []models.MessageEntity{{Type: models.MessageEntityTypeBotCommand, Offset: 0, Length: len("/connect")}},
		},
	}
	b.ProcessUpdate(context.Background(), upd)
	if repo.calls == 0 {
		t.Fatalf("command handler not executed")
	}

	repo.calls = 0
	upd = &models.Update{
		ID: 2,
		CallbackQuery: &models.CallbackQuery{
			ID:      "cb",
			From:    models.User{ID: 1, LanguageCode: "en"},
			Data:    handler.CallbackConnect,
			Message: models.MaybeInaccessibleMessage{Message: &models.Message{ID: 1, Chat: models.Chat{ID: 1}}},
		},
	}
	b.ProcessUpdate(context.Background(), upd)
	if repo.calls == 0 {
		t.Fatalf("callback handler not executed")
	}
}
