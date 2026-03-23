package main

import (
	"log"
	"net/http"

	"infographic-generator/backend/internal/api"
	"infographic-generator/backend/internal/config"
)

func main() {
	cfg := config.Load()
	app := api.New(cfg)

	log.Printf("starting api on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, app.Handler()); err != nil {
		log.Fatal(err)
	}
}
