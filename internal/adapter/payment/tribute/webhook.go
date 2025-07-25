package tribute

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	lru "github.com/hashicorp/golang-lru/v2"

	"remnawave-tg-shop-bot/internal/service/customer"
)

// Handler processes Tribute webhooks.
type Handler struct {
	apiKey      string
	svcCustomer customer.Service
	dedup       *lru.Cache[string, struct{}]
}

// NewHandler constructs Handler with LRU deduplication cache.
func NewHandler(apiKey string, svc customer.Service) *Handler {
	cache, _ := lru.New[string, struct{}](1000)
	return &Handler{apiKey: apiKey, svcCustomer: svc, dedup: cache}
}

type webhookBody struct {
	Event      string `json:"event"`
	Amount     int64  `json:"amount"`
	Currency   string `json:"currency"`
	TelegramID int64  `json:"telegram_id"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	_ = r.Body.Close()

	sig := r.Header.Get("trbt-signature")
	mac := hmac.New(sha256.New, []byte(h.apiKey))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(sig)) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	sum := sha256.Sum256(body)
	key := hex.EncodeToString(sum[:])
	if h.dedup.Contains(key) {
		w.WriteHeader(http.StatusOK)
		return
	}
	h.dedup.Add(key, struct{}{})

	var wb webhookBody
	if err := json.Unmarshal(body, &wb); err != nil {
		slog.Error("webhook: unmarshal", "err", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	switch wb.Event {
	case "new_subscription", "recurrent_donation", "subscription.payment":
		if err := h.svcCustomer.AddBalance(r.Context(), wb.TelegramID, wb.Amount); err != nil {
			slog.Error("add balance", "err", err)
		}
	}

	w.WriteHeader(http.StatusOK)
}
