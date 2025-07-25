package config

import (
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

// Config holds Tribute-specific settings.
type Config struct {
	TributeAPIKey      string
	TributeWebhookPath string
}

type config struct {
	telegramToken                                       string
	price1, price3, price6                              int
	starsPrice1, starsPrice3, starsPrice6               int
	remnawaveUrl, remnawaveToken, remnawaveMode         string
	databaseURL                                         string
	cryptoPayURL, cryptoPayToken                        string
	botURL                                              string
	trafficLimit, trialTrafficLimit                     int
	feedbackURL                                         string
	channelURL                                          string
	serverStatusURL                                     string
	supportURL                                          string
	tosURL                                              string
	isCryptoEnabled                                     bool
	isTelegramStarsEnabled                              bool
	adminTelegramIds                                    map[int64]struct{}
	trialDays                                           int
	inboundUUIDs                                        map[uuid.UUID]uuid.UUID
	referralDays                                        int
	referralBonus                                       int
	miniApp                                             string
	enableAutoPayment                                   bool
	healthCheckPort                                     int
	tributeWebhookUrl, tributeAPIKey, tributePaymentUrl string
	isWebAppLinkEnabled                                 bool
	xApiKey                                             string
	telegramProxyURL                                    string
	telegramProxyHost                                   string
	telegramProxyPort                                   int
	telegramProxyKey                                    string
	telegramProxyChannel                                string
}

var conf config

func GetTributeWebHookUrl() string {
	return conf.tributeWebhookUrl
}
func GetTributeAPIKey() string {
	return conf.tributeAPIKey
}

// Tribute returns Tribute-related configuration.
func Tribute() Config {
	return Config{
		TributeAPIKey:      conf.tributeAPIKey,
		TributeWebhookPath: conf.tributeWebhookUrl,
	}
}

func GetTributePaymentUrl() string {
	return conf.tributePaymentUrl
}

func TelegramProxyURL() string {
	return conf.telegramProxyURL
}

func TelegramProxyHost() string {
	return conf.telegramProxyHost
}

func TelegramProxyPort() int {
	return conf.telegramProxyPort
}

func TelegramProxyKey() string {
	return conf.telegramProxyKey
}

func TelegramProxyChannel() string {
	return conf.telegramProxyChannel
}

func GetReferralDays() int {
	return conf.referralDays
}

func GetReferralBonus() int {
	return conf.referralBonus
}

func GetMiniAppURL() string {
	return conf.miniApp
}

func InboundUUIDs() map[uuid.UUID]uuid.UUID {
	return conf.inboundUUIDs
}

func TrialTrafficLimit() int {
	return conf.trialTrafficLimit * bytesInGigabyte
}

func TrialDays() int {
	return conf.trialDays
}
func FeedbackURL() string {
	return conf.feedbackURL
}

func ChannelURL() string {
	return conf.channelURL
}

func ServerStatusURL() string {
	return conf.serverStatusURL
}

func SupportURL() string {
	return conf.supportURL
}

func TosURL() string {
	return conf.tosURL
}

func Price1() int {
	return conf.price1
}

func Price3() int {
	return conf.price3
}

func Price6() int {
	return conf.price6
}

func Price(month int) int {
	switch month {
	case 1:
		return conf.price1
	case 3:
		return conf.price3
	case 6:
		return conf.price6
	default:
		return conf.price1
	}
}

func StarsPrice(month int) int {
	switch month {
	case 1:
		return conf.starsPrice1
	case 3:
		return conf.starsPrice3
	case 6:
		return conf.starsPrice6
	default:
		return conf.starsPrice1
	}
}
func TelegramToken() string {
	return conf.telegramToken
}
func RemnawaveUrl() string {
	return conf.remnawaveUrl
}
func DatabaseURL() string {
	return conf.databaseURL
}
func RemnawaveToken() string {
	return conf.remnawaveToken
}
func RemnawaveMode() string {
	return conf.remnawaveMode
}
func CryptoPayUrl() string {
	return conf.cryptoPayURL
}
func CryptoPayToken() string {
	return conf.cryptoPayToken
}
func BotURL() string {
	return conf.botURL
}
func SetBotURL(botURL string) {
	conf.botURL = botURL
}
func TrafficLimit() int {
	return conf.trafficLimit * bytesInGigabyte
}

func IsCryptoPayEnabled() bool {
	return conf.isCryptoEnabled
}

func IsTelegramStarsEnabled() bool {
	return conf.isTelegramStarsEnabled
}

func IsAdmin(id int64) bool {
	if conf.adminTelegramIds == nil {
		return false
	}
	_, ok := conf.adminTelegramIds[id]
	return ok
}

func GetAdminTelegramIds() []int64 {
	ids := make([]int64, 0, len(conf.adminTelegramIds))
	for id := range conf.adminTelegramIds {
		ids = append(ids, id)
	}
	return ids
}

func GetHealthCheckPort() int {
	return conf.healthCheckPort
}

func IsWepAppLinkEnabled() bool {
	return conf.isWebAppLinkEnabled
}

func GetXApiKey() string {
	return conf.xApiKey
}

const bytesInGigabyte = 1073741824

func mustEnv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("env %q not set", key)
	}
	return v, nil
}

func mustEnvInt(key string) (int, error) {
	v, err := mustEnv(key)
	if err != nil {
		return 0, err
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid int in %q: %w", key, err)
	}
	return i, nil
}

func envIntDefault(key string, def int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid int in %q: %w", key, err)
	}
	return i, nil
}

func envBool(key string) bool {
	return os.Getenv(key) == "true"
}

func InitConfig() error {
	if os.Getenv("DISABLE_ENV_FILE") != "true" {
		if err := godotenv.Load(".env"); err != nil {
			log.Println("No .env loaded:", err)
		}
	}
	conf.adminTelegramIds = make(map[int64]struct{})
	idsStr := os.Getenv("ADMIN_TELEGRAM_IDS")
	if idsStr == "" {
		return fmt.Errorf("ADMIN_TELEGRAM_IDS .env variable not set")
	}
	for _, idStr := range strings.Split(idsStr, ",") {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid admin id %s: %w", idStr, err)
		}
		conf.adminTelegramIds[id] = struct{}{}
	}

	var err error
	if conf.telegramToken, err = mustEnv("TELEGRAM_TOKEN"); err != nil {
		return err
	}

	conf.xApiKey = os.Getenv("X_API_KEY")

	conf.isWebAppLinkEnabled = func() bool {
		isWebAppLinkEnabled := os.Getenv("IS_WEB_APP_LINK") == "true"
		return isWebAppLinkEnabled
	}()

	conf.miniApp = strings.TrimSpace(os.Getenv("MINI_APP_URL"))

	if conf.trialTrafficLimit, err = mustEnvInt("TRIAL_TRAFFIC_LIMIT"); err != nil {
		return err
	}

	if conf.healthCheckPort, err = envIntDefault("HEALTH_CHECK_PORT", 8080); err != nil {
		return err
	}

	if conf.trialDays, err = mustEnvInt("TRIAL_DAYS"); err != nil {
		return err
	}

	conf.enableAutoPayment = envBool("ENABLE_AUTO_PAYMENT")

	if conf.price1, err = mustEnvInt("PRICE_1"); err != nil {
		return err
	}
	if conf.price3, err = mustEnvInt("PRICE_3"); err != nil {
		return err
	}
	if conf.price6, err = mustEnvInt("PRICE_6"); err != nil {
		return err
	}

	conf.isTelegramStarsEnabled = envBool("TELEGRAM_STARS_ENABLED")
	if conf.isTelegramStarsEnabled {
		if conf.starsPrice1, err = envIntDefault("STARS_PRICE_1", conf.price1); err != nil {
			return err
		}
		if conf.starsPrice3, err = envIntDefault("STARS_PRICE_3", conf.price3); err != nil {
			return err
		}
		if conf.starsPrice6, err = envIntDefault("STARS_PRICE_6", conf.price6); err != nil {
			return err
		}
	}

	if conf.remnawaveUrl, err = mustEnv("REMNAWAVE_URL"); err != nil {
		return err
	}

	switch v := os.Getenv("REMNAWAVE_MODE"); v {
	case "remote", "local":
		conf.remnawaveMode = v
	case "":
		conf.remnawaveMode = "remote"
	default:
		return fmt.Errorf("REMNAWAVE_MODE .env variable must be either 'remote' or 'local'")
	}

	if conf.remnawaveToken, err = mustEnv("REMNAWAVE_TOKEN"); err != nil {
		return err
	}

	if conf.databaseURL, err = mustEnv("DATABASE_URL"); err != nil {
		return err
	}

	conf.isCryptoEnabled = envBool("CRYPTO_PAY_ENABLED")
	if conf.isCryptoEnabled {
		if conf.cryptoPayURL, err = mustEnv("CRYPTO_PAY_URL"); err != nil {
			return err
		}
		if conf.cryptoPayToken, err = mustEnv("CRYPTO_PAY_TOKEN"); err != nil {
			return err
		}
	}

	if conf.trafficLimit, err = mustEnvInt("TRAFFIC_LIMIT"); err != nil {
		return err
	}
	if conf.referralDays, err = envIntDefault("REFERRAL_DAYS", 0); err != nil {
		return err
	}
	if conf.referralBonus, err = envIntDefault("REFERRAL_BONUS", 150); err != nil {
		return err
	}

	conf.serverStatusURL = os.Getenv("SERVER_STATUS_URL")
	conf.supportURL = os.Getenv("SUPPORT_URL")
	conf.feedbackURL = os.Getenv("FEEDBACK_URL")
	conf.channelURL = os.Getenv("CHANNEL_URL")
	conf.tosURL = os.Getenv("TOS_URL")
	conf.telegramProxyURL = os.Getenv("TELEGRAM_PROXY_URL")
	conf.telegramProxyHost = os.Getenv("TELEGRAM_PROXY_HOST")
	if conf.telegramProxyPort, err = envIntDefault("TELEGRAM_PROXY_PORT", 0); err != nil {
		return err
	}
	conf.telegramProxyKey = os.Getenv("TELEGRAM_PROXY_KEY")
	conf.telegramProxyChannel = os.Getenv("TELEGRAM_PROXY_CHANNEL")

	conf.inboundUUIDs = map[uuid.UUID]uuid.UUID{}
	if v := os.Getenv("INBOUND_UUIDS"); v != "" {
		uuids := strings.Split(v, ",")
		for _, value := range uuids {
			uid, pErr := uuid.Parse(value)
			if pErr != nil {
				return fmt.Errorf("parse inbound uuid %s: %w", value, pErr)
			}
			conf.inboundUUIDs[uid] = uid
		}
		slog.Info("Loaded inbound UUIDs", "uuids", uuids)
	} else {
		slog.Info("No inbound UUIDs specified, all will be used")
	}

	conf.tributeWebhookUrl = os.Getenv("TRIBUTE_WEBHOOK_URL")
	if conf.tributeWebhookUrl == "" {
		return fmt.Errorf("TRIBUTE_WEBHOOK_URL is required")
	}
	if conf.tributeAPIKey, err = mustEnv("TRIBUTE_API_KEY"); err != nil {
		return err
	}
	if conf.tributePaymentUrl, err = mustEnv("TRIBUTE_PAYMENT_URL"); err != nil {
		return err
	}

	return ValidateConfig()
}

// ValidateURL verifies if a string is a valid URL with scheme and host.
func ValidateURL(value, name string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is empty", name)
	}
	u, err := url.ParseRequestURI(value)
	if err != nil || u.Scheme == "" || u.Host == "" {
		if err != nil {
			return fmt.Errorf("%s invalid URL %q: %w", name, value, err)
		}
		return fmt.Errorf("%s invalid URL %q", name, value)
	}
	return nil
}

// ValidateToken checks that token is not empty and doesn't contain spaces.
func ValidateToken(value, name string) error {
	v := strings.TrimSpace(value)
	if v == "" {
		return fmt.Errorf("%s is empty", name)
	}
	if strings.ContainsAny(v, " \n\r\t") {
		return fmt.Errorf("%s contains whitespace", name)
	}
	return nil
}

// ValidatePath ensures path starts with '/'.
func ValidatePath(value, name string) error {
	if value == "" {
		return fmt.Errorf("%s is required", name)
	}
	if !strings.HasPrefix(value, "/") {
		return fmt.Errorf("%s must start with \"/\"", name)
	}
	return nil
}

// ValidateConfig validates URLs and tokens loaded from environment.
func ValidateConfig() error {
	if err := ValidateToken(conf.telegramToken, "TELEGRAM_TOKEN"); err != nil {
		return err
	}
	if err := ValidateURL(conf.remnawaveUrl, "REMNAWAVE_URL"); err != nil {
		return err
	}
	if err := ValidateToken(conf.remnawaveToken, "REMNAWAVE_TOKEN"); err != nil {
		return err
	}
	if conf.isCryptoEnabled {
		if err := ValidateURL(conf.cryptoPayURL, "CRYPTO_PAY_URL"); err != nil {
			return err
		}
		if err := ValidateToken(conf.cryptoPayToken, "CRYPTO_PAY_TOKEN"); err != nil {
			return err
		}
	}
	if err := ValidatePath(conf.tributeWebhookUrl, "TRIBUTE_WEBHOOK_URL"); err != nil {
		return err
	}
	if err := ValidateURL(conf.tributePaymentUrl, "TRIBUTE_PAYMENT_URL"); err != nil {
		return err
	}
	if err := ValidateToken(conf.tributeAPIKey, "TRIBUTE_API_KEY"); err != nil {
		return err
	}
	if conf.telegramProxyURL != "" {
		if err := ValidateURL(conf.telegramProxyURL, "TELEGRAM_PROXY_URL"); err != nil {
			return err
		}
	}
	return nil
}
