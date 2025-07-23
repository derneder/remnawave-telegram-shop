package remnawave

import (
	"context"
	"testing"
	"time"

	remapi "github.com/Jolymmiles/remnawave-api-go/api"
	"github.com/google/uuid"

	"remnawave-tg-shop-bot/internal/pkg/contextkey"
)

type stubAPI struct {
	createReq *remapi.CreateUserRequestDto
	updateReq *remapi.UpdateUserRequestDto
}

func (s *stubAPI) UsersControllerGetAllUsers(ctx context.Context, params remapi.UsersControllerGetAllUsersParams, options ...remapi.RequestOption) (*remapi.GetAllUsersResponseDto, error) {
	return nil, nil
}
func (s *stubAPI) UsersControllerGetUserByTelegramId(ctx context.Context, params remapi.UsersControllerGetUserByTelegramIdParams, options ...remapi.RequestOption) (remapi.UsersControllerGetUserByTelegramIdRes, error) {
	return nil, nil
}
func (s *stubAPI) UsersControllerUpdateUser(ctx context.Context, req *remapi.UpdateUserRequestDto, options ...remapi.RequestOption) (*remapi.UserResponseDto, error) {
	s.updateReq = req
	return &remapi.UserResponseDto{Response: remapi.UserDto{}}, nil
}
func (s *stubAPI) InboundsControllerGetInbounds(ctx context.Context, options ...remapi.RequestOption) (*remapi.GetInboundsResponseDto, error) {
	return &remapi.GetInboundsResponseDto{Response: []remapi.GetInboundsResponseDtoResponseItem{{UUID: uuid.New()}}}, nil
}
func (s *stubAPI) UsersControllerCreateUser(ctx context.Context, req *remapi.CreateUserRequestDto, options ...remapi.RequestOption) (*remapi.UserResponseDto, error) {
	s.createReq = req
	return &remapi.UserResponseDto{Response: remapi.UserDto{}}, nil
}
func (s *stubAPI) UsersStatsControllerGetUserUsageByRange(ctx context.Context, params remapi.UsersStatsControllerGetUserUsageByRangeParams, options ...remapi.RequestOption) (remapi.UsersStatsControllerGetUserUsageByRangeRes, error) {
	return nil, nil
}

func TestCreateUserDescription(t *testing.T) {
	api := &stubAPI{}
	c := &Client{client: api}
	ctx := context.WithValue(context.Background(), contextkey.Username, "user")
	if _, err := c.createUser(ctx, 1, 1, 1); err != nil {
		t.Fatalf("createUser: %v", err)
	}
	if !api.createReq.Description.IsSet() {
		t.Fatal("description not set")
	}
	if v, _ := api.createReq.Description.Get(); v != "user" {
		t.Fatalf("expected 'user', got %s", v)
	}

	api.createReq = nil
	if _, err := c.createUser(context.Background(), 1, 1, 1); err != nil {
		t.Fatalf("createUser: %v", err)
	}
	if api.createReq.Description.IsSet() {
		t.Fatal("description should be empty")
	}
}

func TestUpdateUserDescription(t *testing.T) {
	api := &stubAPI{}
	c := &Client{client: api}
	desc := remapi.NilString{}
	desc.SetTo("old")
	existing := &remapi.UserDto{UUID: uuid.New(), ExpireAt: time.Now(), Description: desc}

	ctx := context.WithValue(context.Background(), contextkey.Username, "new")
	if _, err := c.updateUser(ctx, existing, 1, 1); err != nil {
		t.Fatalf("updateUser: %v", err)
	}
	if !api.updateReq.Description.IsSet() {
		t.Fatal("update description not set")
	}
	if v, _ := api.updateReq.Description.Get(); v != "new" {
		t.Fatalf("expected 'new', got %s", v)
	}

	api.updateReq = nil
	ctx = context.WithValue(context.Background(), contextkey.Username, "old")
	if _, err := c.updateUser(ctx, existing, 1, 1); err != nil {
		t.Fatalf("updateUser: %v", err)
	}
	if api.updateReq != nil && api.updateReq.Description.IsSet() {
		t.Fatal("description should not change")
	}
}
