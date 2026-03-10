package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"

	"go-project-2/internal/config"
	"go-project-2/internal/outbox"
	"go-project-2/internal/repository/postgres"
	bookingservice "go-project-2/internal/service"
	httptransport "go-project-2/internal/transport/http"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	store, err := postgres.New(ctx, cfg.PostgresDSN)
	if err != nil {
		logger.Error("init store", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	svc := bookingservice.New(store)
	handler := httptransport.NewHandler(svc)

	kafkaWriter := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.KafkaBrokers...),
		Topic:                  cfg.KafkaTopic,
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           kafka.RequireOne,
		AllowAutoTopicCreation: true,
	}
	defer kafkaWriter.Close()

	worker := outbox.NewWorker(store, kafkaWriter, cfg.OutboxBatchSize, cfg.OutboxPollInterval)
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	go func() {
		if err := worker.Run(workerCtx); err != nil && err != context.Canceled {
			logger.Error("outbox worker stopped", "error", err)
		}
	}()

	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: handler.Router(),
	}

	go func() {
		logger.Info("http server started", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	workerCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("http server shutdown", "error", err)
	}
	time.Sleep(100 * time.Millisecond)
}
