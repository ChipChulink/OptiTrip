package optimizer

import (
	"encoding/json"
	"log"
	"math"
	"sort"

	"github.com/google/uuid"
	"optitrip/internal/core/domain"
)

const (
	MaxRating          = 5.0
	DefaultMaxPlaces   = 10
	DefaultMinutesPace = 360

	WeightInterest   = 0.50
	WeightRating     = 0.20
	WeightPopularity = 0.15
	WeightCostBonus  = 0.10
	WeightTimeBonus  = 0.05

	UtilityBasePerPlace = 10.0
)

type PlaceScore struct {
	Place domain.Place
	Score float64
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
	interests map[string]float64,
	constraints Constraints,
) *OptimizationResult {

	if constraints.MaxPlacesPerDay <= 0 {
		constraints.MaxPlacesPerDay = DefaultMaxPlaces
	}

	catMap := make(map[uuid.UUID]string)
	for _, c := range categories {
		catMap[c.ID] = c.Slug
	}

	filtered := filterPlaces(places, constraints, catMap)
	log.Printf("[DEBUG] Filtered places: %d (from %d)", len(filtered), len(places))

	scored := make([]PlaceScore, 0, len(filtered))
	for _, place := range filtered {
		score := calculateRelevanceScore(place, interests, catMap)
		scored = append(scored, PlaceScore{Place: place, Score: score})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	maxTimePerDay := getMaxTimePerDay(pace)
	selection := optimizeKnapsack(scored, daysCount, budget, maxTimePerDay, constraints, catMap)

	explanation := buildExplanation(selection.Days, constraints)

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
				score += userWeight * pc.Weight * WeightInterest
			}
		}
	}

	score += (place.Rating / MaxRating) * WeightRating
	score += place.PopularityScore * WeightPopularity

	if place.BaseCost > 0 {
		score += (1.0 / math.Log(place.BaseCost+1)) * WeightCostBonus
	}

	if place.AvgDurationMins > 0 {
		score += (1.0 / math.Log(float64(place.AvgDurationMins)+1)) * WeightTimeBonus
	}

	return score
}

func optimizeKnapsack(
	scored []PlaceScore,
	daysCount int,
	totalBudget float64,
	maxTimePerDay int,
	constraints Constraints,
	catMap map[uuid.UUID]string,
) *struct {
	Days      []TripDay
	TotalCost float64
	TotalTime int
} {
	days := make([]TripDay, daysCount)
	for i := 0; i < daysCount; i++ {
		days[i] = TripDay{Day: i + 1, Places: make([]PlaceOutput, 0)}
	}

	bestDays := make([]TripDay, daysCount)
	for i := 0; i < daysCount; i++ {
		bestDays[i] = TripDay{Day: i + 1, Places: make([]PlaceOutput, 0)}
	}

	var bestScore float64 = -1.0
	var finalCost float64
	var finalTime int

	var findOptimal func(placeIdx int, currentCost float64, currentTime int, currentScore float64)
	findOptimal = func(placeIdx int, currentCost float64, currentTime int, currentScore float64) {
		if placeIdx == len(scored) {
			if currentScore > bestScore {
				bestScore = currentScore
				finalCost = currentCost
				finalTime = currentTime
				for i := range days {
					bestDays[i].Places = append([]PlaceOutput(nil), days[i].Places...)
					bestDays[i].TotalCost = days[i].TotalCost
					bestDays[i].TotalTime = days[i].TotalTime
				}
			}
			return
		}

		ps := scored[placeIdx]

		for d := 0; d < daysCount; d++ {
			day := &days[d]

			if len(day.Places) < constraints.MaxPlacesPerDay &&
				day.TotalTime+ps.Place.AvgDurationMins <= maxTimePerDay &&
				currentCost+ps.Place.BaseCost <= totalBudget {

				categoryName := "other"
				if len(ps.Place.PlaceCategories) > 0 {
					if catSlug, ok := catMap[ps.Place.PlaceCategories[0].CategoryID]; ok {
						categoryName = catSlug
					}
				}

				output := PlaceOutput{
					Name:            ps.Place.Name,
					Category:        categoryName,
					AvgDurationMins: ps.Place.AvgDurationMins,
					BaseCost:        ps.Place.BaseCost,
				}

				day.Places = append(day.Places, output)
				day.TotalCost += ps.Place.BaseCost
				day.TotalTime += ps.Place.AvgDurationMins

				findOptimal(placeIdx+1, currentCost+ps.Place.BaseCost, currentTime+ps.Place.AvgDurationMins, currentScore+ps.Score)

				day.Places = day.Places[:len(day.Places)-1]
				day.TotalCost -= ps.Place.BaseCost
				day.TotalTime -= ps.Place.AvgDurationMins
			}
		}

		findOptimal(placeIdx+1, currentCost, currentTime, currentScore)
	}

	maxSearchDepth := len(scored)
	if maxSearchDepth > 25 {
		maxSearchDepth = 25
	}

	findOptimal(0, 0, 0, 0)

	return &struct {
		Days      []TripDay
		TotalCost float64
		TotalTime int
	}{Days: bestDays, TotalCost: finalCost, TotalTime: finalTime}
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
		return DefaultMinutesPace
	}
}

func calculateTotalUtility(days []TripDay) float64 {
	total := 0.0
	for _, day := range days {
		total += float64(len(day.Places)) * UtilityBasePerPlace
	}
	return total
}

func buildExplanation(days []TripDay, constraints Constraints) []string {
	explanation := []string{
		"Маршрут оптимально рассчитан с помощью многомерного алгоритма распределения ограничений",
		"Глобальная функция полезности максимизирована.",
	}
	if constraints.MaxPlacesPerDay > 0 {
		explanation = append(explanation, "Ограничение по количеству активностей строго соблюдено.")
	}
	return explanation
}

func CalculateRouteFromJSON(
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
	return CalculateRoute(places, categories, daysCount, budget, pace, interests, constraints)
}
