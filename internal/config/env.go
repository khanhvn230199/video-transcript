package config

import (
	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

var SvcCfg svcConfig

type svcConfig struct {
	AWS_REGION            string `env:"AWS_REGION" envDefault:""`
	AWS_ACCESS_KEY_ID     string `env:"AWS_ACCESS_KEY_ID" envDefault:""`
	AWS_SECRET_ACCESS_KEY string `env:"AWS_SECRET_ACCESS_KEY" envDefault:""`
	BUCKET_KEY            string `env:"BUCKET_KEY" envDefault:""`
	DURATION_PRE_SIGN_URL int    `env:"DURATION_PRE_SIGN_URL" envDefault:"15"`
	AWS_CUSTOM_ENDPOINT   string `env:"AWS_CUSTOM_ENDPOINT" envDefault:""`
	AWS_BASE_URL          string `env:"AWS_BASE_URL" envDefault:""`

	// App / DB config lấy từ .env
	PostgresUser string `env:"POSTGRES_USER" envDefault:"video_user"`
	PostgresPass string `env:"POSTGRES_PASSWORD" envDefault:"video_password"`
	PostgresDB   string `env:"POSTGRES_DB" envDefault:"video_db"`
	DatabaseURL  string `env:"DATABASE_URL" envDefault:""`

	// JWT config
	JWTSecret       string `env:"JWT_SECRET" envDefault:"changeme-secret"`
	JWTExpireMinute int    `env:"JWT_EXPIRE_MINUTES" envDefault:"60"`
}

func init() {
	log := zap.S().Named("config")
	if err := godotenv.Load(); err != nil {
		log.Errorf("Can not read env from file system, please check the right this program owned.")
	}

	SvcCfg = svcConfig{}

	if err := env.Parse(&SvcCfg); err != nil {
		panic("Can not parse env from file system, please check the env.")
	}
}
