package referral

import referralrepo "remnawave-tg-shop-bot/internal/repository/referral"

// Repository exposes persistence operations for referrals.
type Repository = referralrepo.Repository

// Referral is an alias to the repository Model type.
type Referral = referralrepo.Model
