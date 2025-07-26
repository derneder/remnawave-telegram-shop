package handler

import (
	"context"
	"sync"
	"time"

	"remnawave-tg-shop-bot/internal/pkg/cache"
	"remnawave-tg-shop-bot/internal/pkg/translation"
	pg "remnawave-tg-shop-bot/internal/repository/pg"
	referralrepo "remnawave-tg-shop-bot/internal/repository/referral"
	custrepo "remnawave-tg-shop-bot/internal/service/customer"
	"remnawave-tg-shop-bot/internal/service/payment"
	"remnawave-tg-shop-bot/internal/service/promotion"
	syncsvc "remnawave-tg-shop-bot/internal/service/sync"
)

type PromocodeFinder interface {
	FindByCreator(ctx context.Context, createdBy int64) ([]pg.Promocode, error)
}

type Handler struct {
	customerRepository       custrepo.Repository
	purchaseRepository       *pg.PurchaseRepository
	translation              *translation.Manager
	paymentService           *payment.PaymentService
	syncService              *syncsvc.SyncService
	referralRepository       referralrepo.Repository
	promocodeRepository      PromocodeFinder
	promocodeUsageRepository *pg.PromocodeUsageRepository
	promotionService         promotion.Creator
	cache                    *cache.Cache
	awaitingPromo            map[int64]bool
	awaitingAmount           map[int64]bool
	awaitingCode             map[int64]bool
	awaitingLimit            map[int64]bool
	promoMu                  sync.RWMutex
	fsm                      map[int64]FSMState
	adminStates              map[int64]*adminPromoState
	shortLinks               map[int64][]ShortLink
	shortMu                  sync.RWMutex
}

type ShortLink struct {
	URL       string
	CreatedAt time.Time
}

type FSMState int

const (
	StateNone FSMState = iota
	StateAwaitPromoCode
)

func NewHandler(
	syncService *syncsvc.SyncService,
	paymentService *payment.PaymentService,
	translation *translation.Manager,
	customerRepository custrepo.Repository,
	purchaseRepository *pg.PurchaseRepository,
	referralRepository referralrepo.Repository,
	promocodeRepository PromocodeFinder,
	promocodeUsageRepository *pg.PromocodeUsageRepository,
	promotionService promotion.Creator,
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
		promotionService:         promotionService,
		cache:                    cache,
		awaitingPromo:            make(map[int64]bool),
		awaitingAmount:           make(map[int64]bool),
		awaitingCode:             make(map[int64]bool),
		awaitingLimit:            make(map[int64]bool),
		fsm:                      make(map[int64]FSMState),
		adminStates:              make(map[int64]*adminPromoState),
		shortLinks:               make(map[int64][]ShortLink),
	}
}

func (h *Handler) expectPromo(id int64) {
	h.promoMu.Lock()
	h.awaitingPromo[id] = true
	h.fsm[id] = StateAwaitPromoCode
	h.promoMu.Unlock()
}

func (h *Handler) expectAmount(id int64) {
	h.promoMu.Lock()
	h.awaitingAmount[id] = true
	h.promoMu.Unlock()
}

func (h *Handler) expectCode(id int64) {
	h.promoMu.Lock()
	h.awaitingCode[id] = true
	h.promoMu.Unlock()
}

func (h *Handler) expectLimit(id int64) {
	h.promoMu.Lock()
	h.awaitingLimit[id] = true
	h.promoMu.Unlock()
}

func (h *Handler) consumePromo(id int64) bool {
	h.promoMu.Lock()
	defer h.promoMu.Unlock()
	if h.awaitingPromo[id] || h.fsm[id] == StateAwaitPromoCode {
		delete(h.awaitingPromo, id)
		delete(h.fsm, id)
		return true
	}
	return false
}

func (h *Handler) consumeAmount(id int64) bool {
	h.promoMu.Lock()
	defer h.promoMu.Unlock()
	if h.awaitingAmount[id] {
		delete(h.awaitingAmount, id)
		return true
	}
	return false
}

func (h *Handler) consumeCode(id int64) bool {
	h.promoMu.Lock()
	defer h.promoMu.Unlock()
	if h.awaitingCode[id] {
		delete(h.awaitingCode, id)
		return true
	}
	return false
}

func (h *Handler) consumeLimit(id int64) bool {
	h.promoMu.Lock()
	defer h.promoMu.Unlock()
	if h.awaitingLimit[id] {
		delete(h.awaitingLimit, id)
		return true
	}
	return false
}

func (h *Handler) IsAwaitingPromo(id int64) bool {
	h.promoMu.RLock()
	defer h.promoMu.RUnlock()
	return h.awaitingPromo[id] || h.fsm[id] == StateAwaitPromoCode
}

func (h *Handler) IsAwaitingAmount(id int64) bool {
	h.promoMu.RLock()
	defer h.promoMu.RUnlock()
	return h.awaitingAmount[id]
}

func (h *Handler) IsAwaitingCode(id int64) bool {
	h.promoMu.RLock()
	defer h.promoMu.RUnlock()
	return h.awaitingCode[id]
}

func (h *Handler) IsAwaitingLimit(id int64) bool {
	h.promoMu.RLock()
	defer h.promoMu.RUnlock()
	return h.awaitingLimit[id]
}

func (h *Handler) clearAdminInputs(id int64) {
	h.promoMu.Lock()
	delete(h.awaitingAmount, id)
	delete(h.awaitingCode, id)
	delete(h.awaitingLimit, id)
	h.promoMu.Unlock()
}

const (
	shortLinkTTL       = 5 * time.Minute
	shortLinkMaxStored = 10
)

// Start runs background tasks for handler.
func (h *Handler) Start(ctx context.Context) {
	go h.cleanupShortLinks(ctx)
}

func (h *Handler) cleanupShortLinks(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			h.shortMu.Lock()
			for id, links := range h.shortLinks {
				var filtered []ShortLink
				for _, l := range links {
					if now.Sub(l.CreatedAt) < shortLinkTTL {
						filtered = append(filtered, l)
					}
				}
				if len(filtered) > shortLinkMaxStored {
					filtered = filtered[len(filtered)-shortLinkMaxStored:]
				}
				if len(filtered) == 0 {
					delete(h.shortLinks, id)
				} else {
					h.shortLinks[id] = filtered
				}
			}
			h.shortMu.Unlock()
		case <-ctx.Done():
			return
		}
	}
}
