package config

import (
	"fmt"
	"net/url"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"

	"github.com/cQu1x/Incident-War-Room/internal/errs"
)

type Config struct {
	BotToken string `env:"BOT_TOKEN" env-required:"true"`

	PostgresHost     string `env:"POSTGRES_HOST"     env-default:"localhost"`
	PostgresPort     int    `env:"POSTGRES_PORT"     env-default:"5432"`
	PostgresDB       string `env:"POSTGRES_DB"       env-required:"true"`
	PostgresUser     string `env:"POSTGRES_USER"     env-required:"true"`
	PostgresPassword string `env:"POSTGRES_PASSWORD" env-required:"true"`

	ReportServiceURL string `env:"REPORT_SERVICE_URL" env-default:"http://localhost:8000"`

	TelegraphAccessToken string `env:"TELEGRAPH_ACCESS_TOKEN"`

	S3Enabled       bool   `env:"S3_ENABLED" env-default:"false"`
	S3EndpointURL   string `env:"S3_ENDPOINT_URL"`
	S3Region        string `env:"S3_REGION"`
	S3Bucket        string `env:"S3_BUCKET_NAME"`
	S3AccessKey     string `env:"S3_ACCESS_KEY"`
	S3SecretKey     string `env:"S3_SECRET_KEY"`
	S3PublicBaseURL string `env:"S3_PUBLIC_BASE_URL"`

	HTTPAddr          string `env:"HTTP_ADDR"           env-default:":8080"`
	CORSAllowedOrigin string `env:"CORS_ALLOWED_ORIGIN" env-default:"*"`

	JWTSecret         string        `env:"JWT_SECRET"`
	DashboardURL      string        `env:"DASHBOARD_URL"        env-default:"https://incident-war-room.ru"`
	DashboardTokenTTL time.Duration `env:"DASHBOARD_TOKEN_TTL"  env-default:"168h"`

	AlertChatID              int64  `env:"ALERT_CHAT_ID"`
	AlertmanagerWebhookToken string `env:"ALERTMANAGER_WEBHOOK_TOKEN"`
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, errs.Wrapf(errs.KindInternal, "config.Load", err, "read environment")
	}
	return &cfg, nil
}

// PostgresDSN builds a connection URL for the configured database.
func (c *Config) PostgresDSN() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.PostgresUser, c.PostgresPassword),
		Host:   fmt.Sprintf("%s:%d", c.PostgresHost, c.PostgresPort),
		Path:   c.PostgresDB,
	}
	return u.String()
}
