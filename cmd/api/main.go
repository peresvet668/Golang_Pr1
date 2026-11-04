package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/peresvet668/Golang_Pr1/internal/config"
	"github.com/peresvet668/Golang_Pr1/internal/repository/postgres"
	"github.com/peresvet668/Golang_Pr1/internal/service"
	httpapi "github.com/peresvet668/Golang_Pr1/internal/transport/http"
)

func main() {
	logger := log.New(os.Stdout, "wallet-api ", log.LstdFlags|log.LUTC)
	if err := run(logger); err != nil {
		logger.Fatal(err)
	}
}

func run(logger *log.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	repository, err := postgres.Open(ctx, cfg.DatabaseURL, cfg.MaxOpenConnections)
	if err != nil {
		return err
	}
	defer repository.Close()

	if err := repository.Migrate(ctx); err != nil {
		return err
	}
	if err := repository.EnsureDefaultWallet(ctx); err != nil {
		return err
	}

	server := &http.Server{
		Addr:              cfg.Address,
		Handler:           httpapi.NewHandler(service.New(repository), logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		logger.Printf("HTTP server listening on %s", cfg.Address)
		serveErr <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
