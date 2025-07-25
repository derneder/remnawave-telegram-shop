package tests

import (
	"context"
	"testing"
	"time"

	remapi "github.com/Jolymmiles/remnawave-api-go/api"
	"github.com/google/uuid"

	remnawave "remnawave-tg-shop-bot/internal/adapter/remnawave"
	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	syncsvc "remnawave-tg-shop-bot/internal/service/sync"
)

func TestNewSyncService(t *testing.T) {
	s := syncsvc.NewSyncService(nil, nil)
	if s == nil {
		t.Fatal("nil service")
	}
}

type stubSyncAPI struct{ users []remapi.UserDto }

func (s *stubSyncAPI) UsersControllerGetAllUsers(ctx context.Context, params remapi.UsersControllerGetAllUsersParams, options ...remapi.RequestOption) (*remapi.GetAllUsersResponseDto, error) {
	resp := remapi.GetAllUsersResponseDto{
		Response: remapi.GetAllUsersResponseDtoResponse{
			Users: s.users,
			Total: float64(len(s.users)),
		},
	}
	return &resp, nil
}
func (s *stubSyncAPI) UsersControllerGetUserByTelegramId(ctx context.Context, params remapi.UsersControllerGetUserByTelegramIdParams, options ...remapi.RequestOption) (remapi.UsersControllerGetUserByTelegramIdRes, error) {
	return nil, nil
}
func (s *stubSyncAPI) UsersControllerUpdateUser(ctx context.Context, request *remapi.UpdateUserRequestDto, options ...remapi.RequestOption) (*remapi.UserResponseDto, error) {
	return nil, nil
}
func (s *stubSyncAPI) InboundsControllerGetInbounds(ctx context.Context, options ...remapi.RequestOption) (*remapi.GetInboundsResponseDto, error) {
	return nil, nil
}
func (s *stubSyncAPI) UsersControllerCreateUser(ctx context.Context, request *remapi.CreateUserRequestDto, options ...remapi.RequestOption) (*remapi.UserResponseDto, error) {
	return nil, nil
}
func (s *stubSyncAPI) UsersStatsControllerGetUserUsageByRange(ctx context.Context, params remapi.UsersStatsControllerGetUserUsageByRangeParams, options ...remapi.RequestOption) (remapi.UsersStatsControllerGetUserUsageByRangeRes, error) {
	return nil, nil
}

type stubSyncRepo struct {
	findIDs    []int64
	findRet    []domaincustomer.Customer
	deletedIDs []int64
	created    []domaincustomer.Customer
	updated    []domaincustomer.Customer
}

func (s *stubSyncRepo) FindById(ctx context.Context, id int64) (*domaincustomer.Customer, error) {
	return nil, nil
}
func (s *stubSyncRepo) FindByTelegramId(ctx context.Context, telegramId int64) (*domaincustomer.Customer, error) {
	return nil, nil
}
func (s *stubSyncRepo) Create(ctx context.Context, c *domaincustomer.Customer) (*domaincustomer.Customer, error) {
	return c, nil
}
func (s *stubSyncRepo) UpdateFields(ctx context.Context, id int64, updates map[string]interface{}) error {
	return nil
}
func (s *stubSyncRepo) FindByTelegramIds(ctx context.Context, telegramIDs []int64) ([]domaincustomer.Customer, error) {
	s.findIDs = append([]int64(nil), telegramIDs...)
	return s.findRet, nil
}
func (s *stubSyncRepo) DeleteByNotInTelegramIds(ctx context.Context, telegramIDs []int64) error {
	s.deletedIDs = append([]int64(nil), telegramIDs...)
	return nil
}
func (s *stubSyncRepo) CreateBatch(ctx context.Context, customers []domaincustomer.Customer) error {
	s.created = append(s.created, customers...)
	return nil
}
func (s *stubSyncRepo) UpdateBatch(ctx context.Context, customers []domaincustomer.Customer) error {
	s.updated = append(s.updated, customers...)
	return nil
}
func (s *stubSyncRepo) FindByExpirationRange(ctx context.Context, startDate, endDate time.Time) (*[]domaincustomer.Customer, error) {
	return nil, nil
}

func TestSyncService_NoUsers(t *testing.T) {
	api := &stubSyncAPI{}
	client := remnawave.NewClientWithAPI(api)
	repo := &stubSyncRepo{}
	svc := syncsvc.NewSyncService(client, repo)
	if err := svc.Sync(context.Background()); err == nil {
		t.Fatal("expected error for empty users")
	}
}

func TestSyncService_CreateAndUpdate(t *testing.T) {
	t1 := remapi.NilInt{}
	t1.SetTo(1)
	t2 := remapi.NilInt{}
	t2.SetTo(2)
	users := []remapi.UserDto{
		{UUID: uuid.New(), TelegramId: t1, ExpireAt: time.Now(), SubscriptionUrl: "s1"},
		{UUID: uuid.New(), TelegramId: t2, ExpireAt: time.Now(), SubscriptionUrl: "s2"},
	}
	api := &stubSyncAPI{users: users}
	client := remnawave.NewClientWithAPI(api)
	repo := &stubSyncRepo{findRet: []domaincustomer.Customer{{TelegramID: 1}}}
	svc := syncsvc.NewSyncService(client, repo)
	if err := svc.Sync(context.Background()); err != nil {
		t.Fatalf("sync: %v", err)
	}
	if len(repo.findIDs) != 2 || repo.findIDs[0] != 1 || repo.findIDs[1] != 2 {
		t.Fatalf("unexpected FindByTelegramIds args: %v", repo.findIDs)
	}
	if len(repo.deletedIDs) != 2 {
		t.Fatalf("unexpected delete ids: %v", repo.deletedIDs)
	}
	if len(repo.created) != 1 || repo.created[0].TelegramID != 2 {
		t.Fatalf("expected create for id 2, got %v", repo.created)
	}
	if len(repo.updated) != 1 || repo.updated[0].TelegramID != 1 {
		t.Fatalf("expected update for id 1, got %v", repo.updated)
	}
}
