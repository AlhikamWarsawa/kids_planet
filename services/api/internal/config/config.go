package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Env  string
	Port string

	Postgres PostgresConfig
}

type PostgresConfig struct {
	Host     string
	Port     string
	DB       string
	User     string
	Password string
	SSLMode  string
}

func MustLoad() Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}
	return cfg
}

func Load() (Config, error) {
	cfg := Config{
		Env:  getEnv("ENV", "dev"),
		Port: getEnv("PORT", "8080"),
		Postgres: PostgresConfig{
			Host:     os.Getenv("POSTGRES_HOST"),
			Port:     os.Getenv("POSTGRES_PORT"),
			DB:       os.Getenv("POSTGRES_DB"),
			User:     os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		},
	}

	if err := cfg.Postgres.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c PostgresConfig) Validate() error {
	var missing []string
	if strings.TrimSpace(c.Host) == "" {
		missing = append(missing, "POSTGRES_HOST")
	}
	if strings.TrimSpace(c.Port) == "" {
		missing = append(missing, "POSTGRES_PORT")
	}
	if strings.TrimSpace(c.DB) == "" {
		missing = append(missing, "POSTGRES_DB")
	}
	if strings.TrimSpace(c.User) == "" {
		missing = append(missing, "POSTGRES_USER")
	}
	if strings.TrimSpace(c.Password) == "" {
		missing = append(missing, "POSTGRES_PASSWORD")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required env: %v", missing)
	}
	return nil
}

func (c PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		urlEscape(c.User),
		urlEscape(c.Password),
		c.Host,
		c.Port,
		c.DB,
		c.SSLMode,
	)
}

func getEnv(key, def string) string {
	v := os.Getenv(key)
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

func urlEscape(s string) string {
	repl := map[rune]string{
		'%':  "%25",
		' ':  "%20",
		'!':  "%21",
		'#':  "%23",
		'$':  "%24",
		'&':  "%26",
		'\'': "%27",
		'(':  "%28",
		')':  "%29",
		'*':  "%2A",
		'+':  "%2B",
		',':  "%2C",
		'/':  "%2F",
		':':  "%3A",
		';':  "%3B",
		'=':  "%3D",
		'?':  "%3F",
		'@':  "%40",
		'[':  "%5B",
		']':  "%5D",
	}
	res := ""
	for _, ch := range s {
		if v, ok := repl[ch]; ok {
			res += v
		} else {
			res += string(ch)
		}
	}
	return res
}
