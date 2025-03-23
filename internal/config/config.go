package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	StoragePath string
	Password    string
	SecretKey   string
}

func LoadConfig() *Config {

	config := &Config{}

	if err := godotenv.Load(); err != nil {
		fmt.Printf("no .env file found: %v\n", err)
	}

	port, exists := os.LookupEnv("TODO_PORT")
	if !exists || port == "" {
		fmt.Println("no port in .env, will use default port (:7540)")
		config.Port = ":7540"
	} else {
		if port[0] != ':' {
			port = ":" + port
		}
		config.Port = port
	}

	storagePath, exists := os.LookupEnv("TODO_DBFILE")
	if !exists || storagePath == "" {
		fmt.Println("no path in .env, will use default path (./scheduler.db)")
		config.StoragePath = "./scheduler.db"
	} else {
		config.StoragePath = storagePath
	}

	password, exists := os.LookupEnv("TODO_PASSWORD")
	if !exists || password == "" {
		fmt.Println("no password in .env, must set for auth, will use default password (12345)")
		config.Password = "12345"
	} else {
		config.Password = password
	}

	secretKey, exists := os.LookupEnv("TODO_SECRET")
	if !exists || secretKey == "" {
		fmt.Println("no secret key in .env, must set for auth, will use default secret key (my_secret_key)")
		config.SecretKey = "my_secret_key"
	} else {
		config.SecretKey = secretKey
	}

	return config
}
