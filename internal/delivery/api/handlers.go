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
	tripService    *service.TripService
	cityRepo       domain.CityRepository
	placeRepo      domain.PlaceRepository
	categoryRepo   domain.CategoryRepository
}

func NewHandler(ts *service.TripService, cr domain.CityRepository, pr domain.PlaceRepository, cat domain.CategoryRepository) *Handler {
	return &Handler{
		tripService:  ts,
		cityRepo:     cr,
		placeRepo:    pr,
		categoryRepo: cat,
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
		CityID      string `json:"city_id"`
		DaysCount   int    `json:"days_count"`
		Budget      float64 `json:"budget"`
		Pace        string `json:"pace"`
		Interests   string `json:"interests"`
		Constraints string `json:"constraints"`
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

	log.Printf("[DEBUG] Handler received plan with cost: %.2f, days_json length: %d", plan.TotalCost, len(plan.DaysJSON))

	var days []interface{}
	if plan.DaysJSON != "" {
		if err := json.Unmarshal([]byte(plan.DaysJSON), &days); err != nil {
			log.Printf("[WARNING] Failed to parse days: %v", err)
		}
	}

	var explanation []string
	if plan.ExplanationJSON != "" {
		if err := json.Unmarshal([]byte(plan.ExplanationJSON), &explanation); err != nil {
			log.Printf("[WARNING] Failed to parse explanation: %v", err)
		}
	}

	resp := map[string]interface{}{
		"id":                     plan.ID,
		"total_cost":             plan.TotalCost,
		"total_duration_minutes": plan.TotalDurationMins,
		"utility_score":          plan.UtilityScore,
		"days":                   days,
		"explanation":            explanation,
	}

	log.Printf("[DEBUG] Sending response with %d days", len(days))

	json.NewEncoder(w).Encode(resp)
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

func (h *Handler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categoryRepo.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(categories)
}
