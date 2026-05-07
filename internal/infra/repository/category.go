package repository

import (
	"optitrip/internal/core/domain"
	"github.com/google/uuid"
)

type CategoryRepo struct{}

func NewCategoryRepo() *CategoryRepo {
	return &CategoryRepo{}
}

func (r *CategoryRepo) Create(category *domain.Category) error {
	return DB.Create(category).Error
}

func (r *CategoryRepo) GetByID(id uuid.UUID) (*domain.Category, error) {
	var category domain.Category
	err := DB.First(&category, "id = ?", id).Error
	return &category, err
}

func (r *CategoryRepo) GetBySlug(slug string) (*domain.Category, error) {
	var category domain.Category
	err := DB.First(&category, "slug = ?", slug).Error
	return &category, err
}

func (r *CategoryRepo) List() ([]domain.Category, error) {
	var categories []domain.Category
	err := DB.Find(&categories).Error
	return categories, err
}

func (r *CategoryRepo) Delete(id uuid.UUID) error {
	return DB.Delete(&domain.Category{}, "id = ?", id).Error
}