package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents a set of configurations needed to run the app.
type Config struct {
	Port int `json:"port"`
	Env string `json:"env"`
	Pepper string `json:"pepper"`
	HMACKey string `json:"hmac_key"`
	Database PostgresConfig `json:"database"`
	Github OAuthConfig `json:"github"`
}

// PostgresConfig represents configurations needed to connect to a postgres database.
type PostgresConfig struct {
	Host string `json:"host"`
	Port int `json:"port"`
	User string `json:"user"`
	Password string `json:"password"`
	Name string `json:"name"`
}

// ConnectionInfo returns a PostgresConfig object's values as a string formatted to be
// passed into a method that opens a database connection.
func (pc PostgresConfig) ConnectionInfo() string {
	if pc.Password == "" {
		return fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", pc.Host, pc.Port, pc.User, pc.Name)
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", pc.Host, pc.Port, pc.User, pc.Password, pc.Name)
}

// DefaultConfig returns a Config object populated with default dev environment configuration values.
func DefaultConfig() Config {
	return Config{
		Port:    1111,
		Env:     "dev",
		Pepper:  "secret-random-string",
		HMACKey: "secret-hmac-key",
		Database: DefaultPostgresConfig(),
	}
}

// DefaultPostgresConfig returns a PostgresConfig object populated with default dev environment
// database configuration values.
func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "",
		Name:     "wtf_twitter",
	}
}

// OAuthConfig is a template to hold provider-specific OAuth configuration.
// The actual credentials for each OAuth provider are in .conf.json.
type OAuthConfig struct {
	ID string `json:"id"`
	Secret string `json:"secret"`
	AuthURL string `json:"auth_url"`
	TokenURL string `json:"token_url"`
}

// LoadConfig tries to load production configuration data from a .config.json file,
// decode the data into a Config object and return it. If no config file is found
// it returns the DefaultConfig data meant for dev environments.
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