package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	RefreshIntervalMinutes int    `json:"refresh_interval_minutes"`
	WorkerCount            int    `json:"worker_count"`
	TimeoutSeconds         int    `json:"timeout_seconds"`
	UserAgent              string `json:"user_agent"`
	InputFile              string `json:"input_file"`
	OutputFile             string `json:"output_file"`
}

func LoadConfig() *Config {
	file, _ := os.ReadFile("config.json")
	var cfg Config
	json.Unmarshal(file, &cfg)
	return &cfg
}
