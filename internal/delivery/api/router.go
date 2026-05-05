package api

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

func SetupRouter(h *Handler) *mux.Router {
	r := mux.NewRouter()

	dir, _ := os.Getwd()
	staticDir := filepath.Join(dir, "web", "static")

	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/cities", h.CreateCity).Methods("POST")
	api.HandleFunc("/cities", h.GetCities).Methods("GET")

	api.HandleFunc("/places", h.CreatePlace).Methods("POST")
	api.HandleFunc("/places", h.GetPlaces).Methods("GET")

	api.HandleFunc("/trips/optimize", h.OptimizeTrip).Methods("POST")
	api.HandleFunc("/trips/{id}", h.GetTripPlan).Methods("GET")
	api.HandleFunc("/trips/{id}/recalculate", h.RecalculateTrip).Methods("POST")

	r.HandleFunc("/api/v1/health", h.HealthCheck).Methods("GET")
	r.HandleFunc("/api/v1/metrics", h.Metrics).Methods("GET")

	r.PathPrefix("/").Handler(http.FileServer(http.Dir(staticDir)))

	return r
}