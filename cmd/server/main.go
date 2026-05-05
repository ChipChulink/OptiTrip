package main

import (
	"fmt"
	"log"
	"net/http"

	"optitrip/internal/core/service"
	"optitrip/internal/infra/cache"
	"optitrip/internal/infra/config"
	"optitrip/internal/infra/repository"
	api "optitrip/internal/delivery/api"
)

func main() {
	cfg := config.Load()

	if err := repository.Connect(cfg); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		sqlDB, _ := repository.DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	if err := repository.Migrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	cityRepo := repository.NewCityRepo()
	placeRepo := repository.NewPlaceRepo()
	tripRepo := repository.NewTripRepo()
	appCache := cache.New(cfg.CacheTTL)

	tripService := service.NewTripService(placeRepo, tripRepo, appCache)
	handler := api.NewHandler(tripService, cityRepo, placeRepo)
	router := api.SetupRouter(handler)

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}