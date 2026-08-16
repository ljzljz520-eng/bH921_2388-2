package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Config struct {
	Address    string `json:"address"`
	AdminToken string `json:"admin_token"`
	DataFile   string `json:"data_file"`
}

func Default() Config {
	return Config{
		Address:    "127.0.0.1:8080",
		AdminToken: "admin-fixture-token",
		DataFile:   "var/wedding-live.json",
	}
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	config := Default()
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	if err := config.Validate(); err != nil {
		return Config{}, err
	}
	return config, nil
}

func (c Config) Validate() error {
	if c.Address == "" {
		return errors.New("address is required")
	}
	if c.AdminToken == "" {
		return errors.New("admin token is required")
	}
	if c.DataFile == "" {
		return errors.New("data file is required")
	}
	return nil
}
