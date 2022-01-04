package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Port int `json:"port"`
	Env string `json:"env"`
	Pepper string `json:"pepper"`
	HMACKey string `json:"hmac_key"`
	Database PostgresConfig `json:"database"`
}

type PostgresConfig struct {
	Host string `json:"host"`
	Port int `json:"port"`
	User string `json:"user"`
	Password string `json:"password"`
	Name string `json:"name"`
}

func (pc PostgresConfig) Dialect() string {
	return "postgres"
}

func (pc PostgresConfig) ConnectionInfo() string {
	if pc.Password == "" {
		return fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", pc.Host, pc.Port, pc.User, pc.Name)
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", pc.Host, pc.Port, pc.User, pc.Password, pc.Name)
}

func DefaultConfig() Config {
	return Config{
		Port:    1111,
		Env:     "dev",
		Pepper:  "secret-random-string",
		HMACKey: "secret-hmac-key",
		Database: DefaultPostgresConfig(),
	}
}

func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "",
		Name:     "wtf_twitter",
	}
}

func LoadConfig() Config {
	f, err := os.Open(".config.json")
	if err != nil {
		return DefaultConfig()
	}
	var c Config
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		panic(err)
	}
	fmt.Println("Successfully loaded .config.json")
	return c
}