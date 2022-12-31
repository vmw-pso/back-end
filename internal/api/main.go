package api

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/vmw-pso/back-end/internal/data"
	"github.com/vmw-pso/back-end/internal/jsonlog"

	_ "github.com/lib/pq"
)

const (
	version = "0.0.1"
)

type Config struct {
	Port int64
	Env  string
	DB   struct {
		DSN          string
		MaxOpenConns int
		MaxIdleConns int
		MaxIdleTime  string
	}
	CORS struct {
		TrustedOrigins []string
	}
}

type API struct {
	cfg    *Config
	logger *jsonlog.Logger
	db     *sql.DB
	models data.Models
	wg     sync.WaitGroup
}

func New(cfg *Config, logger *jsonlog.Logger) (*API, error) {
	api := &API{
		cfg:    cfg,
		logger: logger,
	}

	return api, nil
}

func (api *API) Run() error {
	db, err := openDB(api.cfg.DB.DSN, api.cfg.DB.MaxOpenConns, api.cfg.DB.MaxIdleConns, api.cfg.DB.MaxIdleTime)
	if err != nil {
		return err
	}
	defer db.Close()

	api.logger.PrintInfo("database connection pool established", nil)

	api.db = db
	api.models = *data.NewModels(db)

	return api.serve()
}

func openDB(dsn string, maxOpenConns int, maxIdleConns int, maxIdleTime string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)

	duration, err := time.ParseDuration(maxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
