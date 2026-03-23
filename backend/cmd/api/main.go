package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"infographic-generator/backend/internal/api"
	"infographic-generator/backend/internal/config"
	"infographic-generator/backend/internal/modules/projects"
	"infographic-generator/backend/internal/storage"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()

	store, err := projects.NewPostgresStore(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("init postgres store: %v", err)
	}

	blobStorage, err := storage.NewMinIOStorage(ctx, cfg)
	if err != nil {
		store.Close()
		log.Fatalf("init minio storage: %v", err)
	}

	app := api.New(cfg, store, blobStorage)
	defer app.Close()

	log.Printf("starting api on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, app.Handler()); err != nil {
		log.Fatal(err)
	}
}
