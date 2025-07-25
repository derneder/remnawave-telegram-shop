package handler

import "remnawave-tg-shop-bot/internal/pkg/config"

func isAdmin(id int64) bool { return config.IsAdmin(id) }
