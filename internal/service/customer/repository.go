package customer

import domain "remnawave-tg-shop-bot/internal/domain/customer"

// Repository exposes persistence operations for customers.
// It aliases the domain layer interface so services depend only on abstractions.
type Repository = domain.Repository
