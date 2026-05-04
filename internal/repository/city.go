package repository

import (
	"optitrip/internal/domain"
	"github.com/google/uuid"
)

type CityRepo struct{}

func NewCityRepo() *CityRepo {
	return &CityRepo{}
}

func (r *CityRepo) Create(city *domain.City) error {
	return DB.Create(city).Error
}

func (r *CityRepo) GetByID(id uuid.UUID) (*domain.City, error) {
	var city domain.City
	err := DB.First(&city, "id = ?", id).Error
	return &city, err
}

func (r *CityRepo) GetByName(name string) (*domain.City, error) {
	var city domain.City
	err := DB.First(&city, "name = ?", name).Error
	return &city, err
}

func (r *CityRepo) List() ([]domain.City, error) {
	var cities []domain.City
	err := DB.Where("is_active = true").Find(&cities).Error
	return cities, err
}

func (r *CityRepo) Update(city *domain.City) error {
	return DB.Save(city).Error
}

func (r *CityRepo) Delete(id uuid.UUID) error {
	return DB.Delete(&domain.City{}, "id = ?", id).Error
}
