package config

import (
	"log/slog"
	"net/url"
	"pg-bulk-flow/internal/model"
)

type DB struct {
	Addr     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func (cfg DB) ConnectString() string {
	query := url.Values{}
	if cfg.SSLMode != "" {
		query.Set("sslmode", cfg.SSLMode)
	}
	uri := url.URL{
		Scheme:   "postgres",
		Host:     cfg.Addr,
		User:     url.UserPassword(cfg.User, cfg.Password),
		Path:     cfg.Name,
		RawQuery: query.Encode(),
	}
	return uri.String()
}

type Log struct {
	Level     slog.Level
	PlainText bool
}

type Config struct {
	PprofEnable bool
	Log         Log
	DB          DB
	InputFile   string
	NameType    model.NameType
}

func Load() (*Config, error) {
	const required = true
	var ge getenv

	return &Config{
		PprofEnable: ge.Bool("PPROF_ENABLE", !required, false),
		Log: Log{
			Level:     ge.LogLevel("LOG_LEVEL", !required, slog.LevelInfo),
			PlainText: ge.Bool("LOG_PLAINTEXT", !required, false),
		},
		DB: DB{
			Addr:     ge.String("DB_ADDR", required, ""),
			User:     ge.String("DB_USER", !required, "postgres"),
			Password: ge.String("DB_PASSWORD", required, ""),
			Name:     ge.String("DB_NAME", !required, "postgres"),
			SSLMode:  ge.String("DB_SSLMODE", !required, ""),
		},
		InputFile: ge.String("INPUT_FILE", !required, ""),
		NameType:  ge.NameType("NAME_TYPE", !required, model.NameTypeSurname),
	}, ge.Err()
}
