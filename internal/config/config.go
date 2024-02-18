package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	GoroutineCount           int    `json:"goroutine_count"`
	OperationCalculationTime int    `json:"operation_calculation_time"`
	SqlitePath               string `json:"sqlite_path"`
}

func LoadConfig(filename string) (Config, error) {
	cfg := Config{}
	data, err := os.ReadFile(filename)
	if err != nil {
		return cfg, fmt.Errorf("reading config %s failed: %w", filename, err)
	}

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("parsing config failed: %w", err)
	}
	return cfg, nil
}
