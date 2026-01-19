package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Env  string
	Port string

	Postgres PostgresConfig
	JWT      JWTConfig
}

type PostgresConfig struct {
	Host     string
	Port     string
	DB       string
	User     string
	Password string
	SSLMode  string
}

type JWTConfig struct {
	Secret    string
	Issuer    string
	ExpiresIn time.Duration
}

func MustLoad() Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}
	return cfg
}

func Load() (Config, error) {
	jwtExpires, err := parseDurationEnv("JWT_EXPIRES_IN", "24h")
	if err != nil {
		return Config{}, err
	}

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
		JWT: JWTConfig{
			Secret:    os.Getenv("JWT_SECRET"),
			Issuer:    getEnv("JWT_ISSUER", "kids_planet"),
			ExpiresIn: jwtExpires,
		},
	}

	if err := cfg.Postgres.Validate(); err != nil {
		return Config{}, err
	}
	if err := cfg.JWT.Validate(); err != nil {
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

func (c JWTConfig) Validate() error {
	var missing []string
	if strings.TrimSpace(c.Secret) == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if strings.TrimSpace(c.Issuer) == "" {
		missing = append(missing, "JWT_ISSUER")
	}
	if c.ExpiresIn <= 0 {
		return fmt.Errorf("invalid JWT_EXPIRES_IN: must be > 0")
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

func parseDurationEnv(key, def string) (time.Duration, error) {
	raw := getEnv(key, def)
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s=%q (must be Go duration like 12h, 30m, 900s): %w", key, raw, err)
	}
	return d, nil
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
