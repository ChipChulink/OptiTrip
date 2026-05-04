package optimizer

import (
	"encoding/json"
	"math"
	"sort"

	"optitrip/internal/domain"
)

type PlaceScore struct {
	Place     domain.Place
	Score     float64
	DayIndex  int
}

type TripDay struct {
	Day       int
	Places    []domain.Place
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
	constraints := parseConstraints(constraintsJSON)

	filtered := filterPlaces(places, constraints)

	scored := make([]PlaceScore, 0, len(filtered))
	for _, place := range filtered {
		score := calculateRelevanceScore(place, interests)
		scored = append(scored, PlaceScore{Place: place, Score: score})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	selection := greedySelect(scored, daysCount, budget, pace, constraints)

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

func filterPlaces(places []domain.Place, constraints Constraints) []domain.Place {
	filtered := make([]domain.Place, 0)
	for _, place := range places {
		hasExcluded := false
		for _, cat := range constraints.ExcludeCategories {
			if cat == place.Type {
				hasExcluded = true
				break
			}
		}
		if !hasExcluded {
			filtered = append(filtered, place)
		}
	}
	return filtered
}

func calculateRelevanceScore(place domain.Place, interests map[string]float64) float64 {
	score := 0.0

	// Interest match (weight 0.5)
	if weight, ok := interests[place.Type]; ok {
		score += weight * 0.5
	}

	// Rating (weight 0.2)
	score += (place.Rating / 5.0) * 0.2

	// Popularity (weight 0.15)
	score += place.PopularityScore * 0.15

	// Cost efficiency (weight 0.1)
	if place.BaseCost > 0 {
		score += (1.0 / math.Log(place.BaseCost+1)) * 0.1
	}

	// Duration appropriateness (weight 0.05)
	if place.AvgDurationMins > 0 {
		score += (1.0 / math.Log(float64(place.AvgDurationMins)+1)) * 0.05
	}

	return score
}

func greedySelect(scored []PlaceScore, daysCount int, budget float64, pace string, constraints Constraints) *struct {
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

			day.Places = append(day.Places, ps.Place)
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
