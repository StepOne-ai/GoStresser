package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
    DBUser     string
    DBPassword string
    DBName     string
    DBHost     string
    DBPort     string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load(); if err != nil {
		return nil, err
	}

    config := &Config{}

    if dbUser := os.Getenv("DB_USER"); dbUser != "" {
        config.DBUser = dbUser
    } else {
		config.DBUser = "stresser"
	}
    if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
        config.DBPassword = dbPassword
    } else {
		config.DBPassword = "stepan2005"
	}
    if dbName := os.Getenv("DB_NAME"); dbName != "" {
        config.DBName = dbName
    } else {
		config.DBName = "go_stresser"
	}
    if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
        config.DBHost = dbHost
    } else {
		config.DBHost = "localhost"
	}
    if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
        config.DBPort = dbPort
    } else {
		config.DBPort = "5432"
	}

    return config, nil
}