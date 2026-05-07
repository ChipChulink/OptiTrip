package seed

import (
	"log"

	"github.com/google/uuid"
	"optitrip/internal/core/domain"
)

type PlaceData struct {
	Name     string
	Category string
	Cost     float64
	Duration int
	Rating   float64
	Popular  float64
}

type CategoryData struct {
	Name string
	Slug string
}

func Seed(cityRepo domain.CityRepository, categoryRepo domain.CategoryRepository, placeRepo domain.PlaceRepository) error {
	log.Println("Seeding database with extended Smolensk data...")

	city, err := cityRepo.GetByName("Смоленск")
	if err != nil {
		city = &domain.City{
			ID:       uuid.New(),
			Name:     "Смоленск",
			Country:  "Россия",
			IsActive: true,
		}
		if err := cityRepo.Create(city); err != nil {
			return err
		}
		log.Println("City 'Смоленск' created")
	} else {
		log.Println("City 'Смоленск' already exists")
	}

	existingCategories, _ := categoryRepo.List()
	existingMap := make(map[string]uuid.UUID)
	for _, c := range existingCategories {
		existingMap[c.Slug] = c.ID
		log.Printf("Found existing category: %s (%s)", c.Name, c.Slug)
	}

	categoriesToSeed := []CategoryData{
		{"Музеи", "museum"},
		{"Достопримечательности", "landmark"},
		{"Архитектура", "architecture"},
		{"Парки и сады", "park"},
		{"Храмы и церкви", "religious"},
		{"Развлечения", "entertainment"},
		{"Еда и рестораны", "food"},
	}

	categoryMap := make(map[string]uuid.UUID)

	for _, c := range categoriesToSeed {
		if catID, exists := existingMap[c.Slug]; exists {
			categoryMap[c.Slug] = catID
			log.Printf("Category '%s' already exists, skipping", c.Name)
			continue
		}

		cat := &domain.Category{
			ID:   uuid.New(),
			Name: c.Name,
			Slug: c.Slug,
		}
		if err := categoryRepo.Create(cat); err != nil {
			log.Printf("Warning: failed to create category %s: %v", c.Name, err)
			continue
		}
		categoryMap[c.Slug] = cat.ID
		log.Printf("Category '%s' created with ID %s", c.Name, cat.ID)
	}

	placesToSeed := []PlaceData{
		// Museum
		{"Смоленский государственный музей-заповедник", "museum", 500, 180, 4.7, 0.9},
		{"Музей-усадьба М.И. Глинки", "museum", 350, 150, 4.6, 0.8},
		{"Музей «Смоленщина в годы ВОВ»", "museum", 400, 120, 4.8, 0.85},
		{"Смоленский художественный музей", "museum", 300, 120, 4.5, 0.7},
		{"Музей «Русская старина»", "museum", 250, 90, 4.4, 0.65},

		// Landmark
		{"Смоленская крепостная стена", "landmark", 0, 120, 4.9, 0.95},
		{"Памятник Николаю Пржевальскому", "landmark", 0, 30, 4.5, 0.75},
		{"Памятник Софии Ролдугиной", "landmark", 0, 20, 4.3, 0.7},
		{"Памятник Твардовскому и Тёркину", "landmark", 0, 30, 4.6, 0.8},
		{"Сквер Героев", "landmark", 0, 30, 4.4, 0.7},

		// Architecture
		{"Успенский кафедральный собор", "architecture", 0, 60, 4.8, 0.9},
		{"Церковь Иоанна Богослова", "architecture", 0, 30, 4.5, 0.7},
		{"Смоленский драматический театр", "architecture", 800, 180, 4.7, 0.75},
		{"Здание бывшей женской гимназии", "architecture", 0, 30, 4.1, 0.5},
		{"Костел Непорочного Зачатия", "architecture", 0, 45, 4.4, 0.6},

		// Park
		{"Лопатинский сад", "park", 0, 90, 4.7, 0.9},
		{"Парк культуры 1100-летия", "park", 0, 120, 4.5, 0.75},
		{"Реадовский парк", "park", 0, 90, 4.3, 0.65},
		{"Парк Пионеров", "park", 0, 60, 4.2, 0.6},

		// Religious
		{"Свято-Успенский собор", "religious", 0, 60, 4.9, 0.95},
		{"Свято-Троицкий собор", "religious", 0, 45, 4.6, 0.7},
		{"Церковь Спаса Преображения", "religious", 0, 40, 4.5, 0.65},
		{"Храм Михаила Архангела", "religious", 0, 50, 4.4, 0.6},

		// Entertainment
		{"Кинотеатр «Современник»", "entertainment", 450, 150, 4.3, 0.85},
		{"Боулинг-клуб «Смоленск»", "entertainment", 700, 120, 4.2, 0.7},
		{"Квест-комната «Выход есть»", "entertainment", 1500, 90, 4.5, 0.65},
		{"Ночной клуб «Графит»", "entertainment", 1000, 180, 4.0, 0.8},

		// Food
		{"Кафе «КофеСмиль»", "food", 400, 60, 4.6, 0.75},
		{"Ресторан «Губернский»", "food", 1500, 120, 4.4, 0.7},
		{"Пиццерия «ПиццаФактория»", "food", 600, 60, 4.3, 0.8},
		{"Столовая «Домашняя»", "food", 250, 45, 4.2, 0.65},
		{"Кондитерская «Марио»", "food", 350, 40, 4.5, 0.6},
	}

	log.Printf("Creating %d places...", len(placesToSeed))

	for _, p := range placesToSeed {
		place := &domain.Place{
			ID:              uuid.New(),
			CityID:          city.ID,
			Name:            p.Name,
			BaseCost:        p.Cost,
			AvgDurationMins: p.Duration,
			Rating:          p.Rating,
			PopularityScore: p.Popular,
			IsActive:        true,
		}
		if err := placeRepo.Create(place); err != nil {
			log.Printf("Warning: failed to create place %s: %v", p.Name, err)
			continue
		}

		if catID, ok := categoryMap[p.Category]; ok {
			link := domain.PlaceCategory{
				PlaceID:    place.ID,
				CategoryID: catID,
				Weight:    1.0,
			}
			if err := placeRepo.SetCategories(place.ID, []domain.PlaceCategory{link}); err != nil {
				log.Printf("Warning: failed to link place %s to category: %v", p.Name, err)
			}
		} else {
			log.Printf("Warning: category %s not found for place %s", p.Category, p.Name)
		}
	}

	log.Println("Seeding completed successfully")
	return nil
}