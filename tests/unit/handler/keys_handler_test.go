package handler_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	handlerpkg "remnawave-tg-shop-bot/internal/adapter/telegram/handler"
	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	"remnawave-tg-shop-bot/internal/pkg/translation"
	"remnawave-tg-shop-bot/tests/testutils"
)

type documentRecorder struct {
	req  *http.Request
	body []byte
}

func (d *documentRecorder) Do(req *http.Request) (*http.Response, error) {
	d.req = req
	b, _ := io.ReadAll(req.Body)
	d.body = b
	resp := &http.Response{StatusCode: http.StatusOK}
	resp.Body = io.NopCloser(strings.NewReader(`{"ok":true,"result":{"message_id":1}}`))
	resp.Header = make(http.Header)
	resp.Request = req
	return resp, nil
}

func TestKeysCallbackHandler_DecodeBase64(t *testing.T) {
	trans := translation.GetInstance()
	if err := trans.InitDefaultTranslations(); err != nil {
		t.Fatalf("init translations: %v", err)
	}

	encoded := base64.StdEncoding.EncodeToString([]byte("vmess://a\nvmess://b"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(encoded))
	}))
	defer srv.Close()

	link := srv.URL
	repo := &testutils.StubCustomerRepo{CustomerByTelegramID: &domaincustomer.Customer{TelegramID: 1, SubscriptionLink: &link}}

	rec := &documentRecorder{}
	b, err := bot.New("token", bot.WithHTTPClient(time.Second, rec), bot.WithSkipGetMe())
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	h := handlerpkg.NewHandler(nil, nil, trans, repo, nil, nil, nil, nil, nil)

	upd := &models.Update{CallbackQuery: &models.CallbackQuery{From: models.User{ID: 1, LanguageCode: "en"}, Message: models.MaybeInaccessibleMessage{InaccessibleMessage: &models.InaccessibleMessage{Chat: models.Chat{ID: 1}, MessageID: 1}}}}

	h.KeysCallbackHandler(context.Background(), b, upd)

	mediaType, params, err := mime.ParseMediaType(rec.req.Header.Get("Content-Type"))
	if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
		t.Fatalf("unexpected content type: %v", err)
	}
	mr := multipart.NewReader(bytes.NewReader(rec.body), params["boundary"])
	var docData []byte
	for {
		part, errPart := mr.NextPart()
		if errPart == io.EOF {
			break
		}
		if errPart != nil {
			t.Fatalf("read part: %v", errPart)
		}
		if part.FormName() == "document" {
			docData, _ = io.ReadAll(part)
			break
		}
	}
	if string(docData) != "vmess://a\nvmess://b" {
		t.Fatalf("unexpected file content: %q", docData)
	}
}
