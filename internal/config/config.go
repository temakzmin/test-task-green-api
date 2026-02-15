package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig        `mapstructure:"server" validate:"required"`
	CORS      CORSConfig          `mapstructure:"cors" validate:"required"`
	GreenAPI  GreenAPIConfig      `mapstructure:"green_api" validate:"required"`
	Logging   LoggingConfig       `mapstructure:"logging" validate:"required"`
	Validator *validator.Validate `mapstructure:"-"`
}

type ServerConfig struct {
	Host                   string `mapstructure:"host" validate:"required"`
	Port                   int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	ReadTimeoutSeconds     int    `mapstructure:"read_timeout_seconds" validate:"required,min=1"`
	WriteTimeoutSeconds    int    `mapstructure:"write_timeout_seconds" validate:"required,min=1"`
	ShutdownTimeoutSeconds int    `mapstructure:"shutdown_timeout_seconds" validate:"required,min=1"`
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins" validate:"required,min=1,dive,required"`
}

type GreenAPIConfig struct {
	BaseURL        string               `mapstructure:"base_url" validate:"required,url"`
	TimeoutSeconds int                  `mapstructure:"timeout_seconds" validate:"required,min=1"`
	Retry          GreenAPIRetryConfig  `mapstructure:"retry" validate:"required"`
	CircuitBreaker CircuitBreakerConfig `mapstructure:"circuit_breaker" validate:"required"`
}

type GreenAPIRetryConfig struct {
	MaxRetries   int `mapstructure:"max_retries" validate:"min=0,max=10"`
	DelaySeconds int `mapstructure:"delay_seconds" validate:"required,min=1,max=60"`
}

type CircuitBreakerConfig struct {
	Name                string  `mapstructure:"name" validate:"required"`
	ConsecutiveFailures uint32  `mapstructure:"consecutive_failures" validate:"required,min=1,max=50"`
	HalfOpenMaxRequests uint32  `mapstructure:"half_open_max_requests" validate:"required,min=1,max=20"`
	OpenTimeoutSeconds  int     `mapstructure:"open_timeout_seconds" validate:"required,min=1,max=300"`
	IntervalSeconds     int     `mapstructure:"interval_seconds" validate:"required,min=1,max=300"`
	FailureRatio        float64 `mapstructure:"failure_ratio" validate:"required,gte=0,lte=1"`
	MinRequests         uint32  `mapstructure:"min_requests" validate:"required,min=1,max=200"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level" validate:"required,oneof=debug info warn error"`
	Format string `mapstructure:"format" validate:"required,eq=json"`
}

func Load(path string) (Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return Config{}, fmt.Errorf("validate config: %w", err)
	}

	cfg.Validator = validate
	return cfg, nil
}

func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func (s ServerConfig) ReadTimeout() time.Duration {
	return time.Duration(s.ReadTimeoutSeconds) * time.Second
}

func (s ServerConfig) WriteTimeout() time.Duration {
	return time.Duration(s.WriteTimeoutSeconds) * time.Second
}

func (s ServerConfig) ShutdownTimeout() time.Duration {
	return time.Duration(s.ShutdownTimeoutSeconds) * time.Second
}

func (g GreenAPIConfig) Timeout() time.Duration {
	return time.Duration(g.TimeoutSeconds) * time.Second
}

func (r GreenAPIRetryConfig) Delay() time.Duration {
	return time.Duration(r.DelaySeconds) * time.Second
}

func (c CircuitBreakerConfig) OpenTimeout() time.Duration {
	return time.Duration(c.OpenTimeoutSeconds) * time.Second
}

func (c CircuitBreakerConfig) Interval() time.Duration {
	return time.Duration(c.IntervalSeconds) * time.Second
}
