package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/go-yandex-practicum/internal/handlers"
	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/repositories"
	"github.com/sbilibin2017/go-yandex-practicum/internal/services"
)

func run() error {
	err := logger.Initialize(flagLogLevel)
	if err != nil {
		return err
	}

	data := make(map[string]any)

	mfo := repositories.NewMetricFilterOneRepository(data)
	msr := repositories.NewMetricMemorySaveRepository(data)

	service := services.NewMetricUpdateService(mfo, msr)

	handler := handlers.MetricUpdatePathHandler(service)

	router := chi.NewRouter()
	router.Post("/update/{type}/{name}/{value}", handler)
	router.Post("/update/{type}/{name}", handler)

	server := &http.Server{Addr: flagRunAddress, Handler: router}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Log.Infof("Server is running on %s", flagRunAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Errorf("Server error: %v", err)
		}
	}()

	<-ctx.Done()
	logger.Log.Info("Shutting down server...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxShutdown); err != nil {
		logger.Log.Errorf("Shutdown error: %v", err)
	}

	logger.Log.Info("Server stopped gracefully")
	return nil
}
