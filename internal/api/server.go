package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (api *API) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", api.cfg.Port),
		Handler:      api.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		api.logger.PrintInfo("caught signal", map[string]string{
			"signal": s.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		api.logger.PrintInfo("completing background tasks", map[string]string{
			"addr": srv.Addr,
		})

		api.wg.Wait()
		shutdownError <- nil
	}()

	api.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  api.cfg.Env,
	})

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	api.logger.PrintInfo("server stopper", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}
