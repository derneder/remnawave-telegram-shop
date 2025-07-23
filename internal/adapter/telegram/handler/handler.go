package handler

import (
	"sync"
	"time"

	"remnawave-tg-shop-bot/internal/pkg/cache"
	"remnawave-tg-shop-bot/internal/pkg/translation"
	pg "remnawave-tg-shop-bot/internal/repository/pg"
	custrepo "remnawave-tg-shop-bot/internal/service/customer"
	"remnawave-tg-shop-bot/internal/service/payment"
	syncsvc "remnawave-tg-shop-bot/internal/service/sync"
)

type Handler struct {
	customerRepository       custrepo.Repository
	purchaseRepository       *pg.PurchaseRepository
	translation              *translation.Manager
	paymentService           *payment.PaymentService
	syncService              *syncsvc.SyncService
	referralRepository       *pg.ReferralRepository
	promocodeRepository      *pg.PromocodeRepository
	promocodeUsageRepository *pg.PromocodeUsageRepository
	cache                    *cache.Cache
	awaitingPromo            map[int64]bool
	promoMu                  sync.RWMutex
	shortLinks               map[int64][]ShortLink
	shortMu                  sync.RWMutex
}

type ShortLink struct {
	URL       string
	CreatedAt time.Time
}

func NewHandler(
	syncService *syncsvc.SyncService,
	paymentService *payment.PaymentService,
	translation *translation.Manager,
	customerRepository custrepo.Repository,
	purchaseRepository *pg.PurchaseRepository,
	referralRepository *pg.ReferralRepository,
	promocodeRepository *pg.PromocodeRepository,
	promocodeUsageRepository *pg.PromocodeUsageRepository,
	cache *cache.Cache) *Handler {
	return &Handler{
		syncService:              syncService,
		paymentService:           paymentService,
		customerRepository:       customerRepository,
		purchaseRepository:       purchaseRepository,
		translation:              translation,
		referralRepository:       referralRepository,
		promocodeRepository:      promocodeRepository,
		promocodeUsageRepository: promocodeUsageRepository,
		cache:                    cache,
		awaitingPromo:            make(map[int64]bool),
		shortLinks:               make(map[int64][]ShortLink),
	}
}

func (h *Handler) expectPromo(id int64) {
	h.promoMu.Lock()
	h.awaitingPromo[id] = true
	h.promoMu.Unlock()
}

func (h *Handler) consumePromo(id int64) bool {
	h.promoMu.Lock()
	defer h.promoMu.Unlock()
	if h.awaitingPromo[id] {
		delete(h.awaitingPromo, id)
		return true
	}
	return false
}

func (h *Handler) IsAwaitingPromo(id int64) bool {
	h.promoMu.RLock()
	defer h.promoMu.RUnlock()
	return h.awaitingPromo[id]
}
