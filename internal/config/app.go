package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator"
	"github.com/spf13/viper"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
)

type AppConfig struct {
	Password              Password
	JWTSecretKey          string `validate:"required"`
	Email                 Email
	PeriodicCheckDuration time.Duration `validate:"required"`
	Server                Server
	DB                    DB
	AccessBindings        []RoleEndpointBinding
}

type DB struct {
	Host     string `validate:"required"`
	Port     string
	User     string `validate:"required"`
	Password string
	Database string `validate:"required"`
	ShowSql  bool
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
	ServerHost            string        `validate:"required"`
	ServerPort            string        `validate:"required"`
	Password              string        `validate:"required"`
	SenderFromAddress     string        `validate:"required"`
	SenderFromName        string        `validate:"required"`
	SenderWebsiteUrl      string        `validate:"required"`
	ConfirmLinkExpiration time.Duration `validate:"required"`
	IsSendRequired        bool
}

type Password struct {
	ResetLinkExpiration time.Duration `validate:"required"`
	Length              int           `validate:"required,gte=8"`
}

type Server struct {
	Host string `validate:"required"`
	Port int    `validate:"required,gte=1024"`
}

type RoleEndpointBinding struct {
	Role             middlewares.Role
	AllowedEndpoints middlewares.ExistingEndpoints
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
	bindEnvVars()

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
		JWTSecretKey: "default_value",
		DB: DB{
			Host:     "localhost",
			User:     "csr",
			Password: "password",
		},
		Password: Password{
			Length:              8,
			ResetLinkExpiration: 15 * time.Minute,
		},
		Email: Email{
			Password:              "default_value",
			SenderWebsiteUrl:      "https://csr.golangforall.com/",
			ConfirmLinkExpiration: 15 * time.Minute,
		},
		Server: Server{
			Host: "0.0.0.0",
			Port: 8080,
		},
	}
}

func bindEnvVars() {
	viper.BindEnv("jwtsecretkey", "JWT_SECRET_KEY")
	viper.BindEnv("email.password", "EMAIL_PASSWORD")
	viper.BindEnv("db.user", "DB_USER")
	viper.BindEnv("db.password", "DB_PASSWORD")

	viper.AutomaticEnv()
}
