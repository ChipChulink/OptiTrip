package service

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"optitrip/internal/core/domain"
	"optitrip/internal/infra/optimizer"
)

type TripService struct {
	placeRepo     domain.PlaceRepository
	tripRepo      domain.TripRepository
	categoryRepo  domain.CategoryRepository
	cache         domain.Cache
}

func NewTripService(pr domain.PlaceRepository, tr domain.TripRepository, cr domain.CategoryRepository, c domain.Cache) *TripService {
	return &TripService{
		placeRepo:    pr,
		tripRepo:     tr,
		categoryRepo: cr,
		cache:        c,
	}
}

func (s *TripService) CalculateRoute(cityID uuid.UUID, daysCount int, budget float64, pace string, interestsJSON string, constraintsJSON string) (*domain.TripPlan, error) {
	log.Printf("[DEBUG] CalculateRoute started for city: %s, days: %d, budget: %.0f, interests: %s", cityID, daysCount, budget, interestsJSON)

	hash := fmt.Sprintf("%s-%d-%.0f-%s-%s-%s", cityID.String()[:8], daysCount, budget, pace, interestsJSON, constraintsJSON)

	if cached, ok := s.cache.Get(hash); ok {
		log.Println("Cache hit for trip request")
		if plan, ok := cached.(*domain.TripPlan); ok {
			log.Printf("[DEBUG] Cache hit - returning plan with cost: %.2f", plan.TotalCost)
			return plan, nil
		}
		log.Printf("[WARNING] Cache hit but type assertion failed, got %T", cached)
	}

	log.Printf("[DEBUG] Fetching places from DB...")
	places, err := s.placeRepo.GetByCity(cityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get places: %w", err)
	}
	log.Printf("[DEBUG] Got %d places", len(places))

	log.Printf("[DEBUG] Fetching categories...")
	categories, err := s.categoryRepo.List()
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	log.Printf("[DEBUG] Got %d categories", len(categories))

	log.Printf("[DEBUG] Starting optimizer...")
	result := optimizer.CalculateRouteFromJSON(places, categories, daysCount, budget, pace, interestsJSON, constraintsJSON)
	log.Printf("[DEBUG] Optimizer finished, selected %d days", len(result.Days))

	daysJSON, _ := json.Marshal(result.Days)
	explanationJSON, _ := json.Marshal(result.Explanation)

	request := &domain.TripRequest{
		ID:              uuid.New(),
		CityID:          cityID,
		DaysCount:       daysCount,
		Budget:          budget,
		Pace:            pace,
		PreferencesJSON: interestsJSON,
		ConstraintsJSON: constraintsJSON,
		RequestHash:     hash,
	}
	s.tripRepo.CreateRequest(request)

	plan := &domain.TripPlan{
		ID:              uuid.New(),
		TripRequestID:   request.ID,
		TotalCost:       result.TotalCost,
		TotalDurationMins: result.TotalTime,
		UtilityScore:    result.UtilityScore,
		DaysJSON:        string(daysJSON),
		ExplanationJSON: string(explanationJSON),
	}
	s.tripRepo.CreatePlan(plan)

	s.cache.Set(hash, plan)

	return plan, nil
}

func (s *TripService) GetTripPlan(planID uuid.UUID) (*domain.TripPlan, error) {
	return s.tripRepo.GetPlanByID(planID)
}

func (s *TripService) Recalculate(planID uuid.UUID) (*domain.TripPlan, error) {
	plan, err := s.tripRepo.GetPlanByID(planID)
	if err != nil {
		return nil, err
	}

	request, err := s.tripRepo.GetRequestByID(plan.TripRequestID)
	if err != nil {
		return nil, err
	}

	return s.CalculateRoute(request.CityID, request.DaysCount, request.Budget, request.Pace, request.PreferencesJSON, request.ConstraintsJSON)
}

func (s *TripService) ExcludePlaceAndRecalculate(planID uuid.UUID, placeID uuid.UUID) (*domain.TripPlan, error) {

	return s.Recalculate(planID)
}

func (s *TripService) GetCacheStats() (int64, int64) {
	return s.cache.Stats()
}
