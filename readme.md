## Description

A Telegram bot for selling subscriptions with integration to Remnawave (https://remna.st/). This service allows users to
purchase and manage subscriptions through Telegram with multiple payment system options.

- [remnawave-api-go](https://github.com/Jolymmiles/remnawave-api-go)

## Admin commands

- `/sync` - Poll users from remnawave and synchronize them with the database. Remove all users which not present in
  remnawave.
- `/addsubpromo <CODE> <days> <limit>` - create subscription promo for given days and activation limit (0 for infinite).
- `/addbalpromo <amount> <limit>` - create balance promo adding amount (in rubles) with generated code.

### Payment Systems

- [CryptoPay API](https://help.crypt.bot/crypto-pay-api)
- Telegram Stars
- Tribute

### Админские промокоды: подписка и баланс

Администраторы могут выдавать пользователям промокоды двух типов. Кнопки для их создания доступны в разделе «Мои промокоды» только админам.

- **Подписка** — `/addsubpromo <CODE> <days> <limit>` продлевает подписку на указанное количество дней.
- **Баланс** — `/addbalpromo <amount> <limit>` начисляет сумму в рублях и автоматически генерирует код.

#### Telegram Stars pricing

```
⭐ 100 звёзд — 179,00 RUB
⭐ 150 звёзд — 259,00 RUB
⭐ 250 звёзд — 423,99 RUB
⭐ 350 звёзд — 589,00 RUB
⭐ 500 звёзд — 829,00 RUB
⭐ 750 звёзд — 1 239,00 RUB
⭐ 1 000 звёзд — 1 649,00 RUB
⭐ 1 500 звёзд — 2 469,00 RUB
⭐ 2 500 звёзд — 4 099,00 RUB
⭐ 5 000 звёзд — 8 199,00 RUB
⭐ 10 000 звёзд — 16 399,00 RUB
⭐ 25 000 звёзд — 40 999,00 RUB
⭐ 50 000 звёзд — 81 999,00 RUB
⭐ 100 000 звёзд — 163 999,00 RUB
⭐ 150 000 звёзд — 244 999,00 RUB
```

## Features

- Purchase VPN subscriptions with different payment methods (bank cards, cryptocurrency)
- Multiple subscription plans (1, 3, 6 months)
- Automated subscription management
- **Subscription Notifications**: The bot automatically sends notifications to users 3 days before their subscription
  expires, helping them avoid service interruption
- Multi-language support (Russian and English)
- **Selective Inbound Assignment**: Configure specific inbounds to assign to users via UUID filtering
- All telegram message support HTML formatting https://core.telegram.org/bots/api#html-style
- Healthcheck - bot checking availability of db, panel.

## API

Web server start on port defined in .env via HEALTH_CHECK_PORT

- /healthcheck
- /${TRIBUTE_PAYMENT_URL} - webhook for tribute

## Environment Variables

The application requires the following environment variables to be set:
| Variable | Description |
|----------|-------------|
| `PRICE_1` | Price for 1 month |
| `PRICE_3` | Price for 3 months |
| `PRICE_6` | Price for 6 months |
| `HEALTH_CHECK_PORT` | Server port |
| `IS_WEB_APP_LINK` | Show sublink as webapp when true |
| `X_API_KEY` | https://remna.st/docs/security/tinyauth-for-nginx#issuing-api-keys |
| `MINI_APP_URL` | TG Web App URL (optional) |
| `STARS_PRICE_1` | Stars cost for 1 month |
| `STARS_PRICE_3` | Stars cost for 3 months |
| `STARS_PRICE_6` | Stars cost for 6 months |
| `REFERRAL_DAYS` | Referral days (0 to disable) |
| `REFERRAL_BONUS` | Bonus in RUB for successful referral |
| `TELEGRAM_TOKEN` | Telegram Bot API token |
| `DATABASE_URL` | PostgreSQL connection string |
| `POSTGRES_USER` | PostgreSQL username |
| `POSTGRES_PASSWORD` | PostgreSQL password |
| `POSTGRES_DB` | PostgreSQL database name |
| `REMNAWAVE_URL` | Remnawave API URL |
| `REMNAWAVE_MODE` | Remnawave mode remote/local. Use http://remnawave:3000 for local |
| `REMNAWAVE_TOKEN` | Remnawave API authentication token |
| `CRYPTO_PAY_ENABLED` | Enable CryptoPay payments (true/false) |
| `CRYPTO_PAY_TOKEN` | CryptoPay API token |
| `CRYPTO_PAY_URL` | CryptoPay API URL |
| `TRAFFIC_LIMIT` | Max allowed traffic in GB (0 for unlimited) |
| `TELEGRAM_STARS_ENABLED` | Enable Telegram Stars payments (true/false) |
| `SERVER_STATUS_URL` | Optional server status URL |
| `SUPPORT_URL` | Optional support URL |
| `FEEDBACK_URL` | Optional feedback URL |
| `CHANNEL_URL` | Optional Telegram channel URL |
| `ADMIN_TELEGRAM_IDS` | Comma-separated list of admin IDs |
| `TRIAL_TRAFFIC_LIMIT` | Traffic limit for trial subscriptions (GB) |
| `TRIAL_DAYS` | Trial period length in days (0 to disable) |
| `INBOUND_UUIDS` | Comma-separated inbound UUIDs to assign users |
| `TRIBUTE_WEBHOOK_URL` | Path for Tribute webhook handler |
| `TRIBUTE_API_KEY` | Tribute API key |
| `TRIBUTE_PAYMENT_URL` | Tribute subscription link for Telegram |

## User Interface

The bot dynamically creates buttons based on available environment variables:

- Main buttons for purchasing and connecting to the VPN are always shown
- Additional buttons for Server Status, Support, Feedback, and Channel are only displayed if their corresponding URL
  environment variables are set

## Automated Notifications

The bot includes a notification system that runs daily at 16:00 UTC to check for expiring subscriptions:

- Users receive a notification 3 days before their subscription expires
- The notification includes the exact expiration date and a convenient button to renew the subscription
- Notifications are sent in the user's preferred language

## Inbound Configuration

The bot supports selective inbound assignment to users:

- Configure specific inbound UUIDs in the `INBOUND_UUIDS` environment variable (comma-separated)
- If specified, only inbounds with matching UUIDs will be assigned to new users
- If no inbounds match the specified UUIDs or the variable is empty, all available inbounds will be assigned
- This feature allows fine-grained control over which connection methods are available to users

## Plugins and Dependencies

### Telegram Bot

- [Telegram Bot API](https://core.telegram.org/bots/api)
- [Go Telegram Bot API](https://github.com/go-telegram/bot)

### Database

- [PostgreSQL](https://www.postgresql.org/)
- [pgx - PostgreSQL Driver](https://github.com/jackc/pgx)

## Setup Instructions

1. Clone the repository

```bash
git clone https://github.com/Jolymmiels/remnawave-telegram-shop && cd remnawave-telegram-shop
```

2. Create a `.env` file in the root directory with all the environment variables listed above

```bash
mv .env.sample .env
```

3. Запустите контейнеры и примените миграции:

```bash
make dc-up
make dc-migrate
```

## Directory Layout

```
cmd/                Main applications (`bot` and `migrate`)
internal/adapter/   Integrations with external systems (Telegram, Remnawave, payment)
internal/app/       Application initialization and wiring
internal/service/   Business logic
internal/repository/ Data access layer
internal/ui/        Telegram handlers and UI helpers
db/migrations/      SQL migrations
```

### Common Commands

- **Build bot**: `go build ./cmd/bot`
- **Run migrations**: `make dc-migrate`
  Migrations are also executed automatically when the bot starts.
- **Run tests**: `go test ./...`

## Запуск через Docker Compose

Для остановки или просмотра логов доступны удобные цели Makefile:

```bash
make dc-logs    # смотреть логи
make dc-down    # остановить контейнеры
```

## Tribute payment setup instructions

> [!WARNING] 
> To integrate with Tribute, you must have a public domain (e.g., `bot.example.com`) that points to your bot server.  
> Webhook and subscription setup will not work on a local address or IP — only via a domain with a valid SSL certificate.

### How the integration works

The bot supports subscription management via the Tribute service. When a user clicks the payment button, they are redirected to the Tribute bot or payment page to complete the subscription. After successful payment, Tribute sends a webhook to your server, and the bot activates the subscription for the user.

### Step-by-step setup guide

1. Getting started
  * Create a channel;
  * In the Tribute app, open "Channels and Groups" and add your channel;
  * Create a new subscription;
  * Obtain the subscription link (Subscription -> Links -> Telegram Link).

2. Configure environment variables in `.env`
    * Set the webhook path (e.g., `/tribute`):

    ```
    TRIBUTE_WEBHOOK_URL=/tribute
    ```

    * Set the API key from your Tribute settings:

    ```
    TRIBUTE_API_KEY=your_tribute_api_key
    ```

    * Paste the subscription link you got from Tribute:

    ```
    TRIBUTE_PAYMENT_URL=https://t.me/tribute/app?startapp=...
    ```

    * Specify the port the app will use:

    ```
    HEALTH_CHECK_PORT=82251
    ```

3. Restart bot

```bash
docker compose down && docker compose up -d
```

## How to change bot messages

All localized texts are stored in YAML files under the `translations` directory.
Edit `en.yml` and `ru.yml` to update existing keys or add new ones. Run
`make i18n-check` to ensure that both locales contain the same set of keys and
that there are no unused strings in the codebase.

## Update Instructions

1. Pull the latest Docker image:

```bash
docker compose pull
```


2. Restart the containers:

```bash
docker compose down && docker compose up -d
```

## Reverse Proxy Configuration

If you are not using ngrok from `docker-compose.yml`, you need to set up a reverse proxy to forward requests to the bot.

<details>
<summary>Traefik Configuration</summary>
  
```yaml
http:
  routers:
    remnawave-telegram-shop:
      rule: "Host(`bot.example.com`)"
      entrypoints:
        - http
      middlewares:
        - redirect-to-https
      service: remnawave-telegram-shop

    remnawave-telegram-shop-secure:
      rule: "Host(`bot.example.com`)"
      entrypoints:
        - https
      tls:
        certResolver: letsencrypt
      service: remnawave-telegram-shop

  middlewares:
    redirect-to-https:
      redirectScheme:
        scheme: https

  services:
    remnawave-telegram-shop:
      loadBalancer:
        servers:
          - url: "http://bot:82251"
```

</details>

## Development

Run linters and build:

```bash
go vet ./...
staticcheck ./...
golangci-lint run
go build ./cmd/bot
```

Format code before committing:

```bash
gofmt -w .
goimports -w .
```

Start bot locally:

```bash
go run ./cmd/bot
```


## Observability

Prometheus metrics are exposed on the port specified by `HEALTH_CHECK_PORT` (default 8080).
Use Grafana/Prometheus stack to visualize latency and error counters.

## Tests

All tests live under the `tests/` directory.

- `tests/unit` contains unit tests.
- `tests/integration` is reserved for integration tests (run with `-tags integration`).
- `tests/testutils` stores common stubs and helpers used across tests.

Run unit tests with:

```bash
go test ./...
```

Integration tests (if any) can be executed with:

```bash
go test ./tests/integration -tags integration
```

To generate coverage report:

```bash
make cover
```


Ensure `DATABASE_URL` is set to a test PostgreSQL instance before running integration tests.

