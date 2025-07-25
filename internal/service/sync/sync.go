package sync

import (
	"context"
	"fmt"
	"log/slog"
	"remnawave-tg-shop-bot/internal/adapter/remnawave"
	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	custrepo "remnawave-tg-shop-bot/internal/service/customer"
	"time"
)

type SyncService struct {
	client             *remnawave.Client
	customerRepository custrepo.Repository
}

func NewSyncService(client *remnawave.Client, customerRepository custrepo.Repository) *SyncService {
	return &SyncService{
		client: client, customerRepository: customerRepository,
	}
}

func (s SyncService) Sync(ctx context.Context) error {
	slog.Info("Starting sync")
	var telegramIDs []int64
	telegramIDsSet := make(map[int64]int64)
	var mappedUsers []domaincustomer.Customer
	users, err := s.client.GetUsers(ctx)
	if err != nil {
		slog.Error("Error while getting users from remnawave", "err", err)
		return err
	}
	if users == nil || len(*users) == 0 {
		err := fmt.Errorf("no users found in remnawave")
		slog.Error(err.Error())
		return err
	}

	for _, user := range *users {
		if user.TelegramId.Null {
			continue
		}
		if _, exists := telegramIDsSet[int64(user.TelegramId.Value)]; exists {
			continue
		}

		telegramIDsSet[int64(user.TelegramId.Value)] = int64(user.TelegramId.Value)

		telegramIDs = append(telegramIDs, int64(user.TelegramId.Value))

		mappedUsers = append(mappedUsers, domaincustomer.Customer{
			TelegramID:       int64(user.TelegramId.Value),
			ExpireAt:         &user.ExpireAt,
			SubscriptionLink: &user.SubscriptionUrl,
			Balance:          0,
		})
	}

	existingCustomers, err := s.customerRepository.FindByTelegramIds(ctx, telegramIDs)
	if err != nil {
		slog.Error("Error while searching users by telegram ids", "err", err)
		return err
	}
	existingMap := make(map[int64]domaincustomer.Customer)
	for _, cust := range existingCustomers {
		existingMap[cust.TelegramID] = cust
	}

	toCreate := buildToCreate(mappedUsers, existingMap)
	toUpdate := buildToUpdate(mappedUsers, existingMap)

	err = s.customerRepository.DeleteByNotInTelegramIds(ctx, telegramIDs)
	if err != nil {
		slog.Error("Error while deleting users", "err", err)
		return err
	}
	slog.Info("Deleted clients which not exist in panel")

	if len(toCreate) > 0 {
		if err := s.customerRepository.CreateBatch(ctx, toCreate); err != nil {
			slog.Error("Error while creating users", "err", err)
			return err
		}
		slog.Info("Created clients", "count", len(toCreate))
	}

	if len(toUpdate) > 0 {
		if err := s.customerRepository.UpdateBatch(ctx, toUpdate); err != nil {
			slog.Error("Error while updating users", "err", err)
			return err
		}
		slog.Info("Updated clients", "count", len(toUpdate))
	}
	slog.Info("Synchronization completed")
	return nil
}

func buildToCreate(mapped []domaincustomer.Customer, existing map[int64]domaincustomer.Customer) []domaincustomer.Customer {
	var res []domaincustomer.Customer
	for _, cust := range mapped {
		if _, found := existing[cust.TelegramID]; !found {
			res = append(res, cust)
		}
	}
	return res
}

func buildToUpdate(mapped []domaincustomer.Customer, existing map[int64]domaincustomer.Customer) []domaincustomer.Customer {
	var res []domaincustomer.Customer
	for _, cust := range mapped {
		if _, found := existing[cust.TelegramID]; found {
			cust.CreatedAt = time.Now()
			res = append(res, cust)
		}
	}
	return res
}
