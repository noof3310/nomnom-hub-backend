package config

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Server      Server      `envPrefix:"APP_"`
	DB          Database    `envPrefix:"DB_"`
	LineWebhook LineWebhook `envPrefix:"LINE_"`
}

type Server struct {
	Name string `env:"NAME" envDefault:"NomNomHub"`
	Env  string `env:"ENV" envDefault:"dev"`
	Port string `env:"PORT" envDefault:"8080"`
}

type Database struct {
	Host    string `env:"HOST,notEmpty"`
	Port    int    `env:"PORT,notEmpty"`
	User    string `env:"USER,notEmpty"`
	Pass    string `env:"PASS,required"`
	Name    string `env:"NAME,notEmpty"`
	SSLMode string `env:"SSLMODE" envDefault:"disable"`
}

type LineWebhook struct {
	ChannelSecret string `env:"CHANNEL_SECRET,notEmpty"`
	ChannelToken  string `env:"CHANNEL_TOKEN,notEmpty"`
}

func Load() *Config {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("❌ failed to parse environment: %v", err)
	}
	log.Printf("✅ Loaded config for %s [%s]", cfg.Server.Name, cfg.Server.Env)
	return &cfg
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DB.User,
		c.DB.Pass,
		c.DB.Host,
		c.DB.Port,
		c.DB.Name,
		c.DB.SSLMode,
	)
}
