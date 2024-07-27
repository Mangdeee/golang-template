package config

import (
	"fmt"
	"log"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		App
		HTTP
		Log
		PG
		Mail
		AWS
		S3
		Queue
		Monitoring
	}

	App struct {
		Name                  string `env:"APP_NAME"`
		Version               string `env:"APP_VERSION"`
		Url                   string `env:"APP_URL"`
		Secret                string `env:"APP_SECRET"`
		TokenLifespanDuration string `env:"TOKEN_DURATION"`
		TokenLifespan         int    `env:"TOKEN_LIFESPAN"`
		RefreshTokenLifespan  int    `env:"REFRESH_TOKEN_LIFESPAN"`
		DeeplinkUrl           string `env:"DEEPLINK_URL"`
		Timeout               int    `env:"APP_TIMEOUT"`
		Env                   string `env:"APP_ENV"`
	}

	HTTP struct {
		Port string `env:"HTTP_PORT"`
	}

	Log struct {
		Level string `env:"LOGGER_LOG_LEVEL"`
	}

	PG struct {
		PoolMax      int    `env:"PG_POOL_MAX"`
		Host         string `env:"PG_HOST"`
		User         string `env:"PG_USER"`
		Password     string `env:"PG_PASSWORD"`
		DatabaseName string `env:"PG_DBNAME"`
		Port         string `env:"PG_PORT"`
		SslMode      string `env:"PG_SSL_MODE"`
	}

	Mail struct {
		Host     string `env:"MAIL_HOST"`
		Port     int    `env:"MAIL_PORT"`
		User     string `env:"MAIL_USER"`
		Password string `env:"MAIL_PASS"`
		From     string `env:"MAIL_FROM"`
		Test     string `env:"MAIL_TEST"`
	}

	AWS struct {
		Region string `env:"AWS_REGION"`
	}

	S3 struct {
		Bucket string `env:"S3_BUCKET"`
	}

	Queue struct {
		Host string `env:"QUEUE_HOST"`
	}

	Monitoring struct {
		Sentry string `env:"MONITORING_SENTRY"`
	}
)

var (
	once     sync.Once
	instance *Config
)

func GetInstance() *Config {
	if instance == nil {
		once.Do(func() {
			cfg, err := newConfig()

			if err != nil {
				log.Fatal(err)
			}
			instance = cfg
		})
	}

	return instance
}

func newConfig() (*Config, error) {
	cfg := &Config{}
	err := cleanenv.ReadConfig(".env", cfg)
	if err != nil {
		fmt.Println("Using Environment Variable")
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (pg PG) GetDbConnectionUrl() string {
	connectionUrl := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		pg.Host, pg.User, pg.Password, pg.DatabaseName, pg.Port, pg.SslMode,
	)
	return connectionUrl
}
