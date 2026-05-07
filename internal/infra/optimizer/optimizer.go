package optimizer

import (
	"encoding/json"
	"log"
	"math"
	"sort"

	"github.com/google/uuid"
	"optitrip/internal/core/domain"
)

type PlaceScore struct {
	Place     domain.Place
	Score     float64
	DayIndex  int
}

type PlaceOutput struct {
	Name             string  `json:"name"`
	Category         string  `json:"category"`
	AvgDurationMins  int     `json:"avg_duration_minutes"`
	BaseCost         float64 `json:"base_cost"`
}

type TripDay struct {
	Day       int
	Places    []PlaceOutput
	TotalCost float64
	TotalTime int
}

type OptimizationResult struct {
	Days          []TripDay
	TotalCost     float64
	TotalTime     int
	UtilityScore  float64
	Explanation   []string
}

type Interest struct {
	Category string  `json:"category"`
	Weight   float64 `json:"weight"`
}

type Constraints struct {
	MaxPlacesPerDay  int      `json:"max_places_per_day"`
	OnlyOpenPlaces   bool     `json:"only_open_places"`
	ExcludeCategories []string `json:"exclude_categories"`
}

func CalculateRoute(
	places []domain.Place,
	categories []domain.Category,
	daysCount int,
	budget float64,
	pace string,
	interestsJSON string,
	constraintsJSON string,
) *OptimizationResult {

	interests := parseInterests(interestsJSON)
	log.Printf("[DEBUG] Interests parsed: %v", interests)
	constraints := parseConstraints(constraintsJSON)

	catMap := make(map[uuid.UUID]string)
	for _, c := range categories {
		catMap[c.ID] = c.Slug
		log.Printf("[DEBUG] Category ID=%s -> Slug=%s", c.ID, c.Slug)
	}

	if len(places) > 0 && len(places[0].PlaceCategories) > 0 {
		log.Printf("[DEBUG] First place has %d categories", len(places[0].PlaceCategories))
		for _, pc := range places[0].PlaceCategories {
			log.Printf("[DEBUG] PlaceCategory: PlaceID=%s, CategoryID=%s", pc.PlaceID, pc.CategoryID)
		}
	}

	filtered := filterPlaces(places, constraints, catMap)
	log.Printf("[DEBUG] Filtered places: %d (from %d)", len(filtered), len(places))

	scored := make([]PlaceScore, 0, len(filtered))
	for _, place := range filtered {
		score := calculateRelevanceScore(place, interests, catMap)
		scored = append(scored, PlaceScore{Place: place, Score: score})
		if score > 0.5 {
			log.Printf("[DEBUG] Place '%s' score=%.2f (categories: %d)", place.Name, score, len(place.PlaceCategories))
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	selection := greedySelect(scored, daysCount, budget, pace, constraints, catMap)

	explanation := buildExplanation(selection.Days, interests, constraints)

	return &OptimizationResult{
		Days:         selection.Days,
		TotalCost:    selection.TotalCost,
		TotalTime:    selection.TotalTime,
		UtilityScore: calculateTotalUtility(selection.Days),
		Explanation:  explanation,
	}
}

func parseInterests(jsonStr string) map[string]float64 {
	interests := make(map[string]float64)
	if jsonStr == "" {
		return interests
	}
	var list []Interest
	if err := json.Unmarshal([]byte(jsonStr), &list); err == nil {
		for _, i := range list {
			interests[i.Category] = i.Weight
		}
	}
	return interests
}

func parseConstraints(jsonStr string) Constraints {
	c := Constraints{
		MaxPlacesPerDay: 10,
	}
	if jsonStr == "" {
		return c
	}
	json.Unmarshal([]byte(jsonStr), &c)
	return c
}

func filterPlaces(places []domain.Place, constraints Constraints, catMap map[uuid.UUID]string) []domain.Place {
	filtered := make([]domain.Place, 0)
	for _, place := range places {
		hasExcluded := false
		for _, pc := range place.PlaceCategories {
			if catSlug, ok := catMap[pc.CategoryID]; ok {
				for _, excl := range constraints.ExcludeCategories {
					if catSlug == excl {
						hasExcluded = true
						break
					}
				}
			}
			if hasExcluded {
				break
			}
		}
		if !hasExcluded {
			filtered = append(filtered, place)
		}
	}
	return filtered
}

func calculateRelevanceScore(place domain.Place, interests map[string]float64, catMap map[uuid.UUID]string) float64 {
	score := 0.0

	for _, pc := range place.PlaceCategories {
		if catSlug, ok := catMap[pc.CategoryID]; ok {
			if userWeight, ok := interests[catSlug]; ok {
				score += userWeight * pc.Weight * 0.5
			}
		}
	}

	score += (place.Rating / 5.0) * 0.2
	score += place.PopularityScore * 0.15

	if place.BaseCost > 0 {
		score += (1.0 / math.Log(place.BaseCost+1)) * 0.1
	}

	if place.AvgDurationMins > 0 {
		score += (1.0 / math.Log(float64(place.AvgDurationMins)+1)) * 0.05
	}

	return score
}

func greedySelect(scored []PlaceScore, daysCount int, budget float64, pace string, constraints Constraints, catMap map[uuid.UUID]string) *struct {
	Days      []TripDay
	TotalCost float64
	TotalTime int
} {
	days := make([]TripDay, daysCount)
	for i := 0; i < daysCount; i++ {
		days[i] = TripDay{Day: i + 1}
	}

	maxTimePerDay := getMaxTimePerDay(pace)
	used := make(map[string]bool)
	totalCost := 0.0
	totalTime := 0

	for _, ps := range scored {
		if used[ps.Place.ID.String()] {
			continue
		}

		for dayIdx := 0; dayIdx < daysCount; dayIdx++ {
			day := &days[dayIdx]

			if constraints.MaxPlacesPerDay > 0 && len(day.Places) >= constraints.MaxPlacesPerDay {
				continue
			}
			if day.TotalTime+ps.Place.AvgDurationMins > maxTimePerDay {
				continue
			}
			if totalCost+ps.Place.BaseCost > budget {
				continue
			}

			categoryName := "other"
			if len(ps.Place.PlaceCategories) > 0 {
				if catSlug, ok := catMap[ps.Place.PlaceCategories[0].CategoryID]; ok {
					categoryName = catSlug
				}
			}

			placeOutput := PlaceOutput{
				Name:            ps.Place.Name,
				Category:       categoryName,
				AvgDurationMins: ps.Place.AvgDurationMins,
				BaseCost:       ps.Place.BaseCost,
			}

			day.Places = append(day.Places, placeOutput)
			day.TotalCost += ps.Place.BaseCost
			day.TotalTime += ps.Place.AvgDurationMins
			totalCost += ps.Place.BaseCost
			totalTime += ps.Place.AvgDurationMins
			used[ps.Place.ID.String()] = true
			break
		}
	}

	return &struct {
		Days      []TripDay
		TotalCost float64
		TotalTime int
	}{Days: days, TotalCost: totalCost, TotalTime: totalTime}
}

func getMaxTimePerDay(pace string) int {
	switch pace {
	case "relaxed":
		return 240
	case "medium":
		return 360
	case "intensive":
		return 480
	default:
		return 360
	}
}

func calculateTotalUtility(days []TripDay) float64 {
	total := 0.0
	for _, day := range days {
		total += float64(len(day.Places)) * 10.0
	}
	return total
}

func buildExplanation(days []TripDay, interests map[string]float64, constraints Constraints) []string {
	explanation := []string{
		"Маршрут подобран с учетом заданных интересов и ограничений",
		"Максимизирована оценка релевантности маршрута",
	}
	if constraints.MaxPlacesPerDay > 0 {
		explanation = append(explanation, "Соблюдено ограничение по количеству активностей в день")
	}
	return explanation
}
