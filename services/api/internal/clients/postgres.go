package clients

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresConfig struct {
	Host     string
	Port     string
	DB       string
	User     string
	Password string
	SSLMode  string
}

func LoadPostgresConfigFromEnv() PostgresConfig {
	cfg := PostgresConfig{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		DB:       os.Getenv("POSTGRES_DB"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		SSLMode:  os.Getenv("POSTGRES_SSLMODE"),
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}
	return cfg
}

func (c PostgresConfig) Validate() error {
	missing := []string{}
	if c.Host == "" {
		missing = append(missing, "POSTGRES_HOST")
	}
	if c.Port == "" {
		missing = append(missing, "POSTGRES_PORT")
	}
	if c.DB == "" {
		missing = append(missing, "POSTGRES_DB")
	}
	if c.User == "" {
		missing = append(missing, "POSTGRES_USER")
	}
	if c.Password == "" {
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

func NewPostgres(ctx context.Context) (*sql.DB, error) {
	cfg := LoadPostgresConfigFromEnv()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	db, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("sql open: %w", err)
	}

	db.SetMaxOpenConns(15)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("postgres ping failed: %w", err)
	}

	return db, nil
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
	out := make([]rune, 0, len(s))
	escaped := ""
	for _, ch := range s {
		if v, ok := repl[ch]; ok {
			escaped += v
			continue
		}
		out = append(out, ch)
	}
	if escaped == "" {
		return s
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
