package main

import (
	"fmt"
	"log"
	"net/http"

	"optitrip/internal/api"
	"optitrip/internal/cache"
	"optitrip/internal/config"
	"optitrip/internal/repository"
	"optitrip/internal/service"
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
	cache := cache.New(cfg.CacheTTL)

	tripService := service.NewTripService(placeRepo, tripRepo, cache)
	handler := api.NewHandler(tripService, cityRepo, placeRepo)
	router := api.SetupRouter(handler)

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
