package tribute

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	"remnawave-tg-shop-bot/internal/pkg/config"
	pg "remnawave-tg-shop-bot/internal/repository/pg"
	custrepo "remnawave-tg-shop-bot/internal/service/customer"
	"remnawave-tg-shop-bot/internal/service/payment"
	"strings"
	"time"
)

type Client struct {
	paymentService     *payment.PaymentService
	customerRepository custrepo.Repository
}

func NewClient(paymentService *payment.PaymentService, customerRepository custrepo.Repository) *Client {
	return &Client{
		paymentService:     paymentService,
		customerRepository: customerRepository,
	}
}

func (c *Client) WebHookHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Second*60)
		defer cancel()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("webhook: read body error", "error", err)
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		defer func() {
			if cerr := r.Body.Close(); cerr != nil {
				slog.Error("close body", "err", cerr)
			}
		}()

		signature := r.Header.Get("trbt-signature")
		if signature == "" {
			http.Error(w, "missing signature", http.StatusUnauthorized)
			return
		}

		secret := config.GetTributeAPIKey()
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		expected := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(expected), []byte(signature)) {
			slog.Error("webhook: bad signature", "expected", expected)
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		var wh SubscriptionWebhook
		if err := json.Unmarshal(body, &wh); err != nil {
			slog.Error("webhook: unmarshal error", "error", err, "payload", string(body))
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		if wh.Name != "new_subscription" {
			w.WriteHeader(http.StatusOK)
			return
		}

		months := convertPeriodToMonths(wh.Payload.Period)

		customer, err := c.customerRepository.FindByTelegramId(ctx, wh.Payload.TelegramUserID)
		if err != nil {
			slog.Error("webhook: find customer", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if customer == nil {
			customer, err = c.customerRepository.Create(ctx, &domaincustomer.Customer{TelegramID: wh.Payload.TelegramUserID, Language: "ru"})
			if err != nil {
				slog.Error("webhook: create customer", "error", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
		}

		_, purchaseId, err := c.paymentService.CreatePurchase(ctx, wh.Payload.Amount, months, customer, pg.InvoiceTypeTribute)

		if err != nil {
			slog.Error("webhook: create purchase error", "error", err, "payload", string(body))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		err = c.paymentService.ProcessPurchaseById(ctx, purchaseId)
		if err != nil {
			slog.Error("webhook: process purchase error", "error", err, "payload", string(body))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}

func convertPeriodToMonths(period string) int {
	switch strings.ToLower(period) {
	case "monthly":
		return 1
	case "quarterly", "3-month", "3months", "3-months", "q":
		return 3
	case "halfyearly":
		return 6
	default:
		return 1
	}
}
