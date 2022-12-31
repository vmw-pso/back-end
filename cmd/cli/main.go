package main

import (
	"flag"
	"os"
	"strings"

	"github.com/vmw-pso/back-end/internal/api"
	"github.com/vmw-pso/back-end/internal/jsonlog"
)

func main() {
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	if err := run(os.Args, logger); err != nil {
		logger.PrintFatal(err, nil)
	}
}

func run(args []string, logger *jsonlog.Logger) error {
	cfg := api.Config{}

	flags := flag.NewFlagSet(args[0], flag.ExitOnError)

	cfg.Port = *flags.Int64("port", 6543, "port to listen on")
	cfg.Env = *flags.String("env", "development", "executiion environment (development|production)")
	cfg.DB.DSN = *flags.String("db-dsn", "postgres://postgres:password@localhost/pso?sslmode=disable", "database data source name")
	cfg.DB.MaxOpenConns = *flags.Int("db-max-open-conns", 25, "database maximum open connections")
	cfg.DB.MaxIdleConns = *flags.Int("db-max-idle-conns", 25, "database maximum idle connections")
	cfg.DB.MaxIdleTime = *flags.String("db-max-idle-time", "15m", "database maximum idle time")

	flags.Func("cors-trusted-origins", "tructed origins (space separated list)", func(val string) error {
		cfg.CORS.TrustedOrigins = strings.Fields(val)
		return nil
	})

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	api, err := api.New(&cfg, logger)
	if err != nil {
		return err
	}

	return api.Run()
}
