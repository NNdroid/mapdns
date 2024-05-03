package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"mapdns/pkg/common"
)

type Config struct {
	DNS     DNSConfig  `yaml:"dns"`
	Http    HttpConfig `yaml:"http"`
	DB      DBConfig   `yaml:"db"`
	Verbose bool       `yaml:"verbose"`
}

type DNSConfig struct {
	Listen string `yaml:"listen"`
	TTL    uint32 `yaml:"ttl"`
}

type HttpConfig struct {
	Listen string `yaml:"listen"`
}

type DBConfig struct {
	Path string `yaml:"path"`
}

func ReadConfig(path string) (*Config, error) {
	dat, err := common.ReadFile(path)
	if err != nil {
		return &Config{}, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err = yaml.Unmarshal(dat, &cfg); err != nil {
		return &Config{}, fmt.Errorf("failed to decode config: %w", err)
	}

	return &cfg, nil
}
