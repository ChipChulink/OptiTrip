package domain

import (
	"github.com/google/uuid"
)

// PlaceRepository defines the contract for place data access
type PlaceRepository interface {
	Create(place *Place) error
	GetByID(id uuid.UUID) (*Place, error)
	GetByCity(cityID uuid.UUID) ([]Place, error)
	Update(place *Place) error
	Delete(id uuid.UUID) error
	SetCategories(placeID uuid.UUID, categories []PlaceCategory) error
}

// CityRepository defines the contract for city data access
type CityRepository interface {
	Create(city *City) error
	GetByID(id uuid.UUID) (*City, error)
	GetByName(name string) (*City, error)
	List() ([]City, error)
	Update(city *City) error
	Delete(id uuid.UUID) error
}

// TripRepository defines the contract for trip data access
type TripRepository interface {
	CreateRequest(req *TripRequest) error
	GetRequestByID(id uuid.UUID) (*TripRequest, error)
	GetRequestByHash(hash string) (*TripRequest, error)
	CreatePlan(plan *TripPlan) error
	GetPlanByID(id uuid.UUID) (*TripPlan, error)
	GetPlanByRequestID(requestID uuid.UUID) (*TripPlan, error)
	UpdatePlan(plan *TripPlan) error
}

// Cache defines the contract for caching mechanism
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	Delete(key string)
	Clear()
	Stats() (hits, misses int64)
	InvalidateByPrefix(prefix string)
}
