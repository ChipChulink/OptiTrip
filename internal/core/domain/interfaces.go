package domain

import (
	"github.com/google/uuid"
)

type PlaceRepository interface {
	Create(place *Place) error
	GetByID(id uuid.UUID) (*Place, error)
	GetByCity(cityID uuid.UUID) ([]Place, error)
	Update(place *Place) error
	Delete(id uuid.UUID) error
	SetCategories(placeID uuid.UUID, categories []PlaceCategory) error
}

type CityRepository interface {
	Create(city *City) error
	GetByID(id uuid.UUID) (*City, error)
	GetByName(name string) (*City, error)
	List() ([]City, error)
	Update(city *City) error
	Delete(id uuid.UUID) error
}

type CategoryRepository interface {
	Create(category *Category) error
	GetByID(id uuid.UUID) (*Category, error)
	GetBySlug(slug string) (*Category, error)
	List() ([]Category, error)
	Delete(id uuid.UUID) error
}

type TripRepository interface {
	CreateRequest(req *TripRequest) error
	GetRequestByID(id uuid.UUID) (*TripRequest, error)
	GetRequestByHash(hash string) (*TripRequest, error)
	CreatePlan(plan *TripPlan) error
	GetPlanByID(id uuid.UUID) (*TripPlan, error)
	GetPlanByRequestID(requestID uuid.UUID) (*TripPlan, error)
	UpdatePlan(plan *TripPlan) error
}

type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	Delete(key string)
	Clear()
	Stats() (hits, misses int64)
	InvalidateByPrefix(prefix string)
}
