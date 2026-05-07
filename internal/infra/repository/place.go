package repository

import (
	"log"

	"github.com/google/uuid"
	"optitrip/internal/core/domain"
)

type PlaceRepo struct{}

func NewPlaceRepo() *PlaceRepo {
	return &PlaceRepo{}
}

func (r *PlaceRepo) Create(place *domain.Place) error {
	return DB.Create(place).Error
}

func (r *PlaceRepo) GetByID(id uuid.UUID) (*domain.Place, error) {
	var place domain.Place
	err := DB.Preload("PlaceCategories").First(&place, "id = ?", id).Error
	return &place, err
}

func (r *PlaceRepo) GetByCity(cityID uuid.UUID) ([]domain.Place, error) {
	log.Printf("[DEBUG] GetByCity called for city: %s", cityID)
	var places []domain.Place
	err := DB.Preload("PlaceCategories").Where("city_id = ? AND is_active = true", cityID).Find(&places).Error
	
	// Load category info manually for each place
	for i := range places {
		var links []domain.PlaceCategory
		DB.Where("place_id = ?", places[i].ID).Find(&links)
		places[i].PlaceCategories = links
	}
	
	log.Printf("[DEBUG] GetByCity found %d places", len(places))
	return places, err
}

func (r *PlaceRepo) Update(place *domain.Place) error {
	return DB.Save(place).Error
}

func (r *PlaceRepo) Delete(id uuid.UUID) error {
	return DB.Delete(&domain.Place{}, "id = ?", id).Error
}

func (r *PlaceRepo) SetCategories(placeID uuid.UUID, categories []domain.PlaceCategory) error {
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Delete(&domain.PlaceCategory{}, "place_id = ?", placeID).Error; err != nil {
		tx.Rollback()
		return err
	}

	for _, cat := range categories {
		cat.PlaceID = placeID
		if err := tx.Create(&cat).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}