package repository

import (
	"github.com/google/uuid"
	"optitrip/internal/core/domain"
)

type TripRepo struct{}

func NewTripRepo() *TripRepo {
	return &TripRepo{}
}

func (r *TripRepo) CreateRequest(req *domain.TripRequest) error {
	return DB.Create(req).Error
}

func (r *TripRepo) GetRequestByID(id uuid.UUID) (*domain.TripRequest, error) {
	var req domain.TripRequest
	err := DB.First(&req, "id = ?", id).Error
	return &req, err
}

func (r *TripRepo) GetRequestByHash(hash string) (*domain.TripRequest, error) {
	var req domain.TripRequest
	err := DB.First(&req, "request_hash = ?", hash).Error
	return &req, err
}

func (r *TripRepo) CreatePlan(plan *domain.TripPlan) error {
	return DB.Create(plan).Error
}

func (r *TripRepo) GetPlanByID(id uuid.UUID) (*domain.TripPlan, error) {
	var plan domain.TripPlan
	err := DB.First(&plan, "id = ?", id).Error
	return &plan, err
}

func (r *TripRepo) GetPlanByRequestID(requestID uuid.UUID) (*domain.TripPlan, error) {
	var plan domain.TripPlan
	err := DB.First(&plan, "trip_request_id = ?", requestID).Error
	return &plan, err
}

func (r *TripRepo) UpdatePlan(plan *domain.TripPlan) error {
	return DB.Save(plan).Error
}
