# OptiTrip - Сервис оптимизации туристических маршрутов
Backend-приложение на Go для автоматического подбора оптимального туристического маршрута.

## Быстрый старт
### Настройка базы данных
1. Создайте базу данных `optitrip`
2. Скопируйте `configs/.env.example` в `configs/.env` и настройте параметры подключения

### Запуск
```bash
go run cmd/server/main.go
```

## Пример использования
### 1. Создание города
```bash
curl -X POST http://localhost:8080/api/v1/cities \
  -H "Content-Type: application/json" \
  -d '{"name":"Санкт-Петербург","country":"Россия"}'
```

### 2. Добавление достопримечательности
```bash
curl -X POST http://localhost:8080/api/v1/places \
  -H "Content-Type: application/json" \
  -d '{"city_id":"<ID_ГОРОДА>","name":"Эрмитаж","type":"museum","base_cost":1000,"avg_duration_minutes":180,"rating":4.8,"popularity_score":0.95,"latitude":59.9398,"longitude":30.3146}'
```

### 3. Расчет маршрута
```bash
curl -X POST http://localhost:8080/api/v1/trips/optimize \
  -H "Content-Type: application/json" \
  -d '{
    "city_id": "<ID_ГОРОДА>",
    "days_count": 3,
    "budget": 15000,
    "pace": "medium",
    "interests": [{"category":"museum","weight":0.9}],
    "constraints": {"max_places_per_day": 4}
  }'
```

## Структура проекта
```
/cmd/server         - точка входа
/internal/api       - HTTP-слой
/internal/domain    - доменные сущности
/internal/service   - прикладные сервисы
/internal/optimizer - модуль оптимизации
/internal/repository - работа с БД
/internal/cache     - кеширование
/internal/config    - конфигурация
```

## Основные эндпоинты
- `POST /api/v1/cities` - создать город
- `GET /api/v1/cities` - список городов
- `POST /api/v1/places` - создать достопримечательность
- `GET /api/v1/places?city_id=...` - список мест
- `POST /api/v1/trips/optimize` - рассчитать маршрут
- `GET /api/v1/trips/{id}` - получить маршрут
- `GET /api/v1/health` - проверка состояния
