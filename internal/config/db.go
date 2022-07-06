package config

import "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"

type DB struct {
	Host string
}

func NewDBConfig() *DB {
	dbHost := utils.GetEnv("DB_HOST", "localhost")
	return &DB{
		Host: dbHost,
	}
}
