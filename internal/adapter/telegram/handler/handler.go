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

type Handler struct {
	customerRepository       custrepo.Repository
	purchaseRepository       *pg.PurchaseRepository
	translation              *translation.Manager
	paymentService           *payment.PaymentService
	syncService              *syncsvc.SyncService
	referralRepository       referralrepo.Repository
	promocodeRepository      *pg.PromocodeRepository
	promocodeUsageRepository *pg.PromocodeUsageRepository
	promotionService         promotion.Creator
	cache                    *cache.Cache
	awaitingPromo            map[int64]bool
	promoMu                  sync.RWMutex
	adminStates              map[int64]*adminPromoState
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
	referralRepository referralrepo.Repository,
	promocodeRepository *pg.PromocodeRepository,
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
		adminStates:              make(map[int64]*adminPromoState),
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
