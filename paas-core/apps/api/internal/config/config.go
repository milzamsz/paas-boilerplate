package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds the entire application configuration.
type Config struct {
	App        AppConfig        `mapstructure:"app" yaml:"app"`
	Database   DatabaseConfig   `mapstructure:"database" yaml:"database"`
	JWT        JWTConfig        `mapstructure:"jwt" yaml:"jwt"`
	Server     ServerConfig     `mapstructure:"server" yaml:"server"`
	Logging    LoggingConfig    `mapstructure:"logging" yaml:"logging"`
	Ratelimit  RateLimitConfig  `mapstructure:"ratelimit" yaml:"ratelimit"`
	Migrations MigrationsConfig `mapstructure:"migrations" yaml:"migrations"`
	Health     HealthConfig     `mapstructure:"health" yaml:"health"`
	Xendit     XenditConfig     `mapstructure:"xendit" yaml:"xendit"`
	CORS       CORSConfig       `mapstructure:"cors" yaml:"cors"`
	Email      EmailConfig      `mapstructure:"email" yaml:"email"`
	Storage    StorageConfig    `mapstructure:"storage" yaml:"storage"`
	OAuth      OAuthConfig      `mapstructure:"oauth" yaml:"oauth"`
	Supabase   SupabaseConfig   `mapstructure:"supabase" yaml:"supabase"`
}

type AppConfig struct {
	Name        string `mapstructure:"name" yaml:"name"`
	Version     string `mapstructure:"version" yaml:"version"`
	Environment string `mapstructure:"environment" yaml:"environment"`
	Debug       bool   `mapstructure:"debug" yaml:"debug"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host" yaml:"host"`
	Port     int    `mapstructure:"port" yaml:"port"`
	User     string `mapstructure:"user" yaml:"user"`
	Password string `mapstructure:"password" yaml:"password"`
	Name     string `mapstructure:"name" yaml:"name"`
	SSLMode  string `mapstructure:"sslmode" yaml:"sslmode"`
}

// DSN returns the PostgreSQL connection string.
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

type JWTConfig struct {
	Secret          string        `mapstructure:"secret" yaml:"secret"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl" yaml:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl" yaml:"refresh_token_ttl"`
}

type ServerConfig struct {
	Port            string `mapstructure:"port" yaml:"port"`
	ReadTimeout     int    `mapstructure:"readtimeout" yaml:"readtimeout"`
	WriteTimeout    int    `mapstructure:"writetimeout" yaml:"writetimeout"`
	IdleTimeout     int    `mapstructure:"idletimeout" yaml:"idletimeout"`
	ShutdownTimeout int    `mapstructure:"shutdowntimeout" yaml:"shutdowntimeout"`
	MaxHeaderBytes  int    `mapstructure:"maxheaderbytes" yaml:"maxheaderbytes"`
}

type LoggingConfig struct {
	Level string `mapstructure:"level" yaml:"level"`
}

type RateLimitConfig struct {
	Enabled  bool          `mapstructure:"enabled" yaml:"enabled"`
	Requests int           `mapstructure:"requests" yaml:"requests"`
	Window   time.Duration `mapstructure:"window" yaml:"window"`
}

type MigrationsConfig struct {
	Directory   string `mapstructure:"directory" yaml:"directory"`
	Timeout     int    `mapstructure:"timeout" yaml:"timeout"`
	LockTimeout int    `mapstructure:"locktimeout" yaml:"locktimeout"`
}

type HealthConfig struct {
	Timeout              int  `mapstructure:"timeout" yaml:"timeout"`
	DatabaseCheckEnabled bool `mapstructure:"database_check_enabled" yaml:"database_check_enabled"`
}

type XenditConfig struct {
	SecretKey    string `mapstructure:"secret_key" yaml:"secret_key"`
	WebhookToken string `mapstructure:"webhook_token" yaml:"webhook_token"`
	CallbackURL  string `mapstructure:"callback_url" yaml:"callback_url"`
}

type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins" yaml:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods" yaml:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers" yaml:"allowed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials" yaml:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age" yaml:"max_age"`
}

type EmailConfig struct {
	Provider  string `mapstructure:"provider" yaml:"provider"`     // "resend" (default), "smtp"
	APIKey    string `mapstructure:"api_key" yaml:"api_key"`       // Resend API key
	FromEmail string `mapstructure:"from_email" yaml:"from_email"` // e.g. "PaaS <noreply@example.com>"
	AppURL    string `mapstructure:"app_url" yaml:"app_url"`       // e.g. "https://app.example.com" for link generation
}

// StorageConfig configures S3-compatible storage providers.
// Works with AWS S3, MinIO, Storj, IDrive e2, Cloudflare R2, etc.
type StorageConfig struct {
	Endpoint        string `mapstructure:"endpoint" yaml:"endpoint"`                   // e.g. "https://gateway.storjshare.io" or "https://e2.idrivee2.com"
	Region          string `mapstructure:"region" yaml:"region"`                       // e.g. "us-east-1"
	Bucket          string `mapstructure:"bucket" yaml:"bucket"`                       // bucket name
	AccessKeyID     string `mapstructure:"access_key_id" yaml:"access_key_id"`         // S3 access key
	SecretAccessKey string `mapstructure:"secret_access_key" yaml:"secret_access_key"` // S3 secret key
	UsePathStyle    bool   `mapstructure:"use_path_style" yaml:"use_path_style"`       // true for MinIO/Storj/IDrive
	PublicURL       string `mapstructure:"public_url" yaml:"public_url"`               // optional CDN/custom URL prefix
}

// OAuthConfig configures external OAuth identity providers.
type OAuthConfig struct {
	Google      OAuthProviderConfig `mapstructure:"google" yaml:"google"`
	GitHub      OAuthProviderConfig `mapstructure:"github" yaml:"github"`
	FrontendURL string              `mapstructure:"frontend_url" yaml:"frontend_url"` // redirect target after callback
}

// OAuthProviderConfig holds credentials for a single OAuth provider.
type OAuthProviderConfig struct {
	ClientID     string `mapstructure:"client_id" yaml:"client_id"`
	ClientSecret string `mapstructure:"client_secret" yaml:"client_secret"`
	Enabled      bool   `mapstructure:"enabled" yaml:"enabled"`
}

// SupabaseConfig configures Supabase integration (cloud or community on-prem).
type SupabaseConfig struct {
	Enabled       bool   `mapstructure:"enabled" yaml:"enabled"`               // master switch for Supabase auth
	URL           string `mapstructure:"url" yaml:"url"`                       // e.g. https://xyz.supabase.co or http://localhost:8000
	AnonKey       string `mapstructure:"anon_key" yaml:"anon_key"`             // public anon key
	ServiceKey    string `mapstructure:"service_key" yaml:"service_key"`       // service_role key (admin operations)
	JWTSecret     string `mapstructure:"jwt_secret" yaml:"jwt_secret"`         // for HS256 JWT validation
	DBHost        string `mapstructure:"db_host" yaml:"db_host"`               // direct DB host (optional, for GORM)
	DBPort        int    `mapstructure:"db_port" yaml:"db_port"`               // direct DB port (default 5432)
	DBName        string `mapstructure:"db_name" yaml:"db_name"`               // database name (default "postgres")
	DBUser        string `mapstructure:"db_user" yaml:"db_user"`               // database user (default "postgres")
	DBPassword    string `mapstructure:"db_password" yaml:"db_password"`       // database password
	DBSSLMode     string `mapstructure:"db_sslmode" yaml:"db_sslmode"`         // SSL mode (default "disable" for local)
	WebhookSecret string `mapstructure:"webhook_secret" yaml:"webhook_secret"` // secret to verify Supabase webhook payloads
}

// DSN returns the Supabase PostgreSQL connection string.
func (s *SupabaseConfig) DSN() string {
	host := s.DBHost
	if host == "" {
		host = "localhost"
	}
	port := s.DBPort
	if port == 0 {
		port = 5432
	}
	dbName := s.DBName
	if dbName == "" {
		dbName = "postgres"
	}
	user := s.DBUser
	if user == "" {
		user = "postgres"
	}
	sslMode := s.DBSSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, s.DBPassword, dbName, sslMode,
	)
}

// LoadConfig loads configuration using Viper. If configPath is non-empty it
// will be used as the exact config file path, otherwise Viper searches common locations.
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	bindEnvVariables(v)

	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}
		}
	} else {
		env := v.GetString("APP_ENVIRONMENT")
		if env == "" {
			env = "development"
		}

		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("configs")
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")

		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("failed to read base config file: %w", err)
			}
		}

		// Merge environment-specific config (e.g. config.development.yaml)
		v.SetConfigName(fmt.Sprintf("config.%s", env))
		if err := v.MergeInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("failed to merge environment config: %w", err)
			}
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Fallback for environment
	if cfg.App.Environment == "" {
		if e := v.GetString("app.environment"); e != "" {
			cfg.App.Environment = e
		} else if e := v.GetString("APP_ENVIRONMENT"); e != "" {
			cfg.App.Environment = e
		} else {
			cfg.App.Environment = "development"
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks required config values.
func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if c.JWT.Secret == "" || c.JWT.Secret == "change-me-in-production" {
		if c.App.Environment == "production" {
			return fmt.Errorf("JWT secret must be set in production")
		}
	}
	return nil
}

func bindEnvVariables(v *viper.Viper) {
	envBindings := map[string]string{
		"app.name":                      "APP_NAME",
		"app.version":                   "APP_VERSION",
		"app.environment":               "APP_ENVIRONMENT",
		"app.debug":                     "APP_DEBUG",
		"database.host":                 "DATABASE_HOST",
		"database.port":                 "DATABASE_PORT",
		"database.user":                 "DATABASE_USER",
		"database.password":             "DATABASE_PASSWORD",
		"database.name":                 "DATABASE_NAME",
		"database.sslmode":              "DATABASE_SSLMODE",
		"jwt.secret":                    "JWT_SECRET",
		"jwt.access_token_ttl":          "JWT_ACCESS_TOKEN_TTL",
		"jwt.refresh_token_ttl":         "JWT_REFRESH_TOKEN_TTL",
		"server.port":                   "SERVER_PORT",
		"server.readtimeout":            "SERVER_READTIMEOUT",
		"server.writetimeout":           "SERVER_WRITETIMEOUT",
		"logging.level":                 "LOGGING_LEVEL",
		"ratelimit.enabled":             "RATELIMIT_ENABLED",
		"ratelimit.requests":            "RATELIMIT_REQUESTS",
		"ratelimit.window":              "RATELIMIT_WINDOW",
		"migrations.directory":          "MIGRATIONS_DIRECTORY",
		"health.database_check_enabled": "HEALTH_DATABASE_CHECK_ENABLED",
		"xendit.secret_key":             "XENDIT_SECRET_KEY",
		"xendit.webhook_token":          "XENDIT_WEBHOOK_TOKEN",
		"xendit.callback_url":           "XENDIT_CALLBACK_URL",
		"cors.allowed_origins":          "CORS_ALLOWED_ORIGINS",
		"cors.allow_credentials":        "CORS_ALLOW_CREDENTIALS",
		"oauth.google.client_id":        "OAUTH_GOOGLE_CLIENT_ID",
		"oauth.google.client_secret":    "OAUTH_GOOGLE_CLIENT_SECRET",
		"oauth.google.enabled":          "OAUTH_GOOGLE_ENABLED",
		"oauth.github.client_id":        "OAUTH_GITHUB_CLIENT_ID",
		"oauth.github.client_secret":    "OAUTH_GITHUB_CLIENT_SECRET",
		"oauth.github.enabled":          "OAUTH_GITHUB_ENABLED",
		"oauth.frontend_url":            "OAUTH_FRONTEND_URL",
		// Supabase
		"supabase.enabled":        "SUPABASE_ENABLED",
		"supabase.url":            "SUPABASE_URL",
		"supabase.anon_key":       "SUPABASE_ANON_KEY",
		"supabase.service_key":    "SUPABASE_SERVICE_KEY",
		"supabase.jwt_secret":     "SUPABASE_JWT_SECRET",
		"supabase.db_host":        "SUPABASE_DB_HOST",
		"supabase.db_port":        "SUPABASE_DB_PORT",
		"supabase.db_name":        "SUPABASE_DB_NAME",
		"supabase.db_user":        "SUPABASE_DB_USER",
		"supabase.db_password":    "SUPABASE_DB_PASSWORD",
		"supabase.db_sslmode":     "SUPABASE_DB_SSLMODE",
		"supabase.webhook_secret": "SUPABASE_WEBHOOK_SECRET",
	}
	for key, env := range envBindings {
		_ = v.BindEnv(key, env)
	}
}

// GetLogLevel converts the string log level to slog.Level.
func (l *LoggingConfig) GetLogLevel() slog.Level {
	switch strings.ToLower(l.Level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// GetSkipPaths returns paths to exclude from request logging.
func GetSkipPaths(env string) []string {
	switch env {
	case "production":
		return []string{"/health", "/health/live", "/health/ready", "/metrics"}
	default:
		return []string{"/health", "/health/live", "/health/ready"}
	}
}

// GetConfigPath searches common config file locations.
func GetConfigPath() string {
	paths := []string{
		"configs/config.yaml",
		"./configs/config.yaml",
		"../configs/config.yaml",
	}
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return "configs/config.yaml"
}

// LogSafeConfig prints configuration to the logger with secrets redacted.
func (c *Config) LogSafeConfig(logger *slog.Logger) {
	logger.Info("Loaded Configuration:")
	logger.Info("App", "Name", c.App.Name, "Environment", c.App.Environment, "Debug", c.App.Debug)
	logger.Info("Database", "Host", c.Database.Host, "Port", c.Database.Port, "Name", c.Database.Name, "SSLMode", c.Database.SSLMode)
	logger.Info("JWT", "Secret", "<redacted>", "AccessTokenTTL", c.JWT.AccessTokenTTL, "RefreshTokenTTL", c.JWT.RefreshTokenTTL)
	logger.Info("Server", "Port", c.Server.Port, "ReadTimeout", c.Server.ReadTimeout, "WriteTimeout", c.Server.WriteTimeout)
	logger.Info("RateLimit", "Enabled", c.Ratelimit.Enabled, "Requests", c.Ratelimit.Requests, "Window", c.Ratelimit.Window)
	logger.Info("OAuth", "GoogleEnabled", c.OAuth.Google.Enabled, "GitHubEnabled", c.OAuth.GitHub.Enabled, "FrontendURL", c.OAuth.FrontendURL)
	logger.Info("Supabase", "Enabled", c.Supabase.Enabled, "URL", c.Supabase.URL, "AnonKey", "<redacted>", "ServiceKey", "<redacted>")
}
