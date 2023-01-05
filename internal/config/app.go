package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator"
	"github.com/spf13/viper"
)

type AppConfig struct {
	Password                            Password
	JWTSecretKey                        string `validate:"required"`
	Email                               Email
	OrderStatusOverdueTimeCheckDuration time.Duration `validate:"required"`
	Server                              Server
	DB                                  DB
}

type DB struct {
	Host          string `validate:"required"`
	Port          string
	User          string `validate:"required"`
	Password      string
	Database      string `validate:"required"`
	EntMigrations bool
}

func (db DB) GetConnectionString() string {
	s := fmt.Sprintf("host=%s", db.Host)
	if db.Port != "" {
		s = fmt.Sprintf("%s port=%s", s, db.Port)
	}
	if db.User != "" {
		s = fmt.Sprintf("%s user=%s", s, db.User)
	}
	if db.Database != "" {
		s = fmt.Sprintf("%s database=%s", s, db.Database)
	}
	if db.Password != "" {
		s = fmt.Sprintf("%s password=%s", s, db.Password)
	}

	return s
}

type Email struct {
	ServerHost        string `validate:"required"`
	ServerPort        string `validate:"required"`
	Password          string `validate:"required"`
	SenderFromAddress string `validate:"required"`
	SenderFromName    string `validate:"required"`
	SenderWebsiteUrl  string `validate:"required"`
	IsSendRequired    bool
}

type Password struct {
	ResetExpirationMinutes time.Duration `validate:"required"`
	Length                 int           `validate:"required,gte=8"`
}

type Server struct {
	Host string `validate:"required"`
	Port int    `validate:"required,gte=1024"`
}

func GetAppConfig(additionalDirectories ...string) (*AppConfig, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	for _, d := range additionalDirectories {
		viper.AddConfigPath(d)
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read in config: %w", err)
	}

	conf := getDefaultConfig()
	if err := viper.Unmarshal(&conf); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config into struct: %w", err)
	}

	if err := validator.New().Struct(conf); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return conf, nil
}

func getDefaultConfig() *AppConfig {
	return &AppConfig{
		DB: DB{
			Host:     "localhost",
			User:     "csr",
			Password: "password",
		},
		Password: Password{
			Length:                 8,
			ResetExpirationMinutes: 15 * time.Minute,
		},
		Email: Email{
			SenderWebsiteUrl: "https://csr.golangforall.com/",
		},
		Server: Server{
			Host: "127.0.0.1",
			Port: 8080,
		},
	}
}
