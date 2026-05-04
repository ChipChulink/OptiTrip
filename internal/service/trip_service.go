package service

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"optitrip/internal/cache"
	"optitrip/internal/domain"
	"optitrip/internal/optimizer"
	"optitrip/internal/repository"
)

type TripService struct {
	placeRepo *repository.PlaceRepo
	tripRepo  *repository.TripRepo
	cache     *cache.Cache
}

func NewTripService(pr *repository.PlaceRepo, tr *repository.TripRepo, c *cache.Cache) *TripService {
	return &TripService{
		placeRepo: pr,
		tripRepo:  tr,
		cache:     c,
	}
}

func (s *TripService) CalculateRoute(cityID uuid.UUID, daysCount int, budget float64, pace string, interestsJSON string, constraintsJSON string) (*domain.TripPlan, error) {

	hash := fmt.Sprintf("%s-%d-%.2f-%s-%s-%s", cityID, daysCount, budget, pace, interestsJSON, constraintsJSON)

	if cached, ok := s.cache.Get(hash); ok {
		log.Println("Cache hit for trip request")
		if plan, ok := cached.(*domain.TripPlan); ok {
			return plan, nil
		}
	}

	places, err := s.placeRepo.GetByCity(cityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get places: %w", err)
	}

	result := optimizer.CalculateRoute(places, nil, daysCount, budget, pace, interestsJSON, constraintsJSON)

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
