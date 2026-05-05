package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"optitrip/internal/core/domain"
	"optitrip/internal/core/service"
)

type Handler struct {
	tripService *service.TripService
	cityRepo    domain.CityRepository
	placeRepo   domain.PlaceRepository
}

func NewHandler(ts *service.TripService, cr domain.CityRepository, pr domain.PlaceRepository) *Handler {
	return &Handler{
		tripService: ts,
		cityRepo:    cr,
		placeRepo:   pr,
	}
}

func (h *Handler) CreateCity(w http.ResponseWriter, r *http.Request) {
	var city domain.City
	if err := json.NewDecoder(r.Body).Decode(&city); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	city.ID = uuid.New()
	if err := h.cityRepo.Create(&city); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(city)
}

func (h *Handler) GetCities(w http.ResponseWriter, r *http.Request) {
	cities, err := h.cityRepo.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(cities)
}

func (h *Handler) CreatePlace(w http.ResponseWriter, r *http.Request) {
	var place domain.Place
	if err := json.NewDecoder(r.Body).Decode(&place); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	place.ID = uuid.New()
	if err := h.placeRepo.Create(&place); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(place)
}

func (h *Handler) GetPlaces(w http.ResponseWriter, r *http.Request) {
	cityIDStr := r.URL.Query().Get("city_id")
	if cityIDStr == "" {
		http.Error(w, "city_id required", http.StatusBadRequest)
		return
	}
	cityID, err := uuid.Parse(cityIDStr)
	if err != nil {
		http.Error(w, "Invalid city_id", http.StatusBadRequest)
		return
	}
	places, err := h.placeRepo.GetByCity(cityID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(places)
}

func (h *Handler) OptimizeTrip(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CityID          string `json:"city_id"`
		DaysCount       int    `json:"days_count"`
		Budget          float64 `json:"budget"`
		Pace            string `json:"pace"`
		Interests       string `json:"interests"`
		Constraints     string `json:"constraints"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	cityID, err := uuid.Parse(req.CityID)
	if err != nil {
		http.Error(w, "Invalid city_id", http.StatusBadRequest)
		return
	}

	plan, err := h.tripService.CalculateRoute(cityID, req.DaysCount, req.Budget, req.Pace, req.Interests, req.Constraints)
	if err != nil {
		log.Printf("Error calculating route: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(plan)
}

func (h *Handler) GetTripPlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid plan ID", http.StatusBadRequest)
		return
	}

	plan, err := h.tripService.GetTripPlan(planID)
	if err != nil {
		http.Error(w, "Plan not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(plan)
}

func (h *Handler) RecalculateTrip(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid plan ID", http.StatusBadRequest)
		return
	}

	plan, err := h.tripService.Recalculate(planID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(plan)
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) Metrics(w http.ResponseWriter, r *http.Request) {
	hits, misses := h.tripService.GetCacheStats()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"cache_hits":   hits,
		"cache_misses": misses,
	})
}
