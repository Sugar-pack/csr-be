package config

import (
	"strconv"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
)

type Server struct {
	Host string
	Port int
}

func SetupServerConfig() (*Server, error) {
	serverHost := utils.GetEnv("SERVER_HOST", "0.0.0.0")
	serverPort := utils.GetEnv("SERVER_PORT", "8080")

	serverPortInt, err := strconv.Atoi(serverPort)
	if err != nil {
		return nil, err
	}

	return &Server{
		Host: serverHost,
		Port: serverPortInt,
	}, nil
}
