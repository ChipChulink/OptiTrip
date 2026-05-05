package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type City struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string         `gorm:"size:255;not null" json:"name"`
	Country   string         `gorm:"size:100" json:"country"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Category struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Slug      string    `gorm:"size:100;unique;not null" json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

type Place struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	CityID           uuid.UUID  `gorm:"type:uuid;not null;index" json:"city_id"`
	Name             string     `gorm:"size:255;not null" json:"name"`
	Description      string     `gorm:"type:text" json:"description"`
	Type             string     `gorm:"size:50" json:"type"`
	BaseCost         float64    `gorm:"default:0" json:"base_cost"`
	AvgDurationMins  int        `gorm:"column:avg_duration_minutes;default:60" json:"avg_duration_minutes"`
	Rating           float64    `gorm:"default:0" json:"rating"`
	PopularityScore  float64    `gorm:"default:0" json:"popularity_score"`
	Latitude         float64    `gorm:"default:0" json:"latitude"`
	Longitude        float64    `gorm:"default:0" json:"longitude"`
	IsActive         bool       `gorm:"default:true" json:"is_active"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	Categories       []Category `gorm:"many2many:place_categories;" json:"categories,omitempty"`
	PlaceCategories  []PlaceCategory `gorm:"foreignKey:PlaceID" json:"-"`
}

type PlaceCategory struct {
	PlaceID  uuid.UUID `gorm:"type:uuid;primaryKey" json:"place_id"`
	CategoryID uuid.UUID `gorm:"type:uuid;primaryKey" json:"category_id"`
	Weight   float64    `gorm:"default:1" json:"weight"`
}

type PlaceSchedule struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	PlaceID   uuid.UUID `gorm:"type:uuid;not null;index" json:"place_id"`
	Weekday   int       `gorm:"not null" json:"weekday"`
	OpenTime  string    `gorm:"size:5" json:"open_time"`
	CloseTime string    `gorm:"size:5" json:"close_time"`
	IsClosed  bool      `gorm:"default:false" json:"is_closed"`
}

type TripRequest struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CityID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"city_id"`
	DaysCount       int        `gorm:"not null" json:"days_count"`
	Budget          float64    `gorm:"not null" json:"budget"`
	Pace            string     `gorm:"size:20" json:"pace"`
	PreferencesJSON string     `gorm:"type:text" json:"preferences_json"`
	ConstraintsJSON string     `gorm:"type:text" json:"constraints_json"`
	RequestHash     string     `gorm:"size:64;index" json:"request_hash"`
	CreatedAt       time.Time  `json:"created_at"`
}

type TripPlan struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TripRequestID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"trip_request_id"`
	TotalCost       float64   `gorm:"default:0" json:"total_cost"`
	TotalDurationMins int     `gorm:"column:total_duration_minutes;default:0" json:"total_duration_minutes"`
	UtilityScore    float64   `gorm:"default:0" json:"utility_score"`
	DaysJSON        string    `gorm:"type:text" json:"days_json"`
	ExplanationJSON string    `gorm:"type:text" json:"explanation_json"`
	CreatedAt       time.Time `json:"created_at"`
}

type CatalogVersion struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CityID    uuid.UUID `gorm:"type:uuid;not null;index" json:"city_id"`
	Version   int       `gorm:"default:1" json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
}
