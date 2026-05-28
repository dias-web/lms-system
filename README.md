# LMS Main Service

Это бэкенд для системы управления обучением (Learning Management System), который я делал в рамках стажировки в BITLAB ACADEMY.

Сервис умеет хранить курсы, главы и уроки и отдавать их по REST API. Внутри — PostgreSQL, миграции через Goose, документация в Swagger.

## Стек

- Go 1.25
- Gin — HTTP-фреймворк
- GORM — ORM поверх PostgreSQL
- Goose — миграции
- logrus — логирование
- Testify + Mockery — юнит-тесты
- swag — генерация Swagger-доки из аннотаций
- Docker + Docker Compose

## Быстрый старт (через Docker Hub)

Образ уже выложен на Docker Hub, локально ничего собирать не надо.

```bash
# 1. Клонируем репо (нужен только docker-compose.yml и .env.example)
git clone https://github.com/dias-web/lms-system.git
cd lms-system

# 2. Копируем переменные окружения
cp .env.example .env

# 3. Поднимаем стек
docker compose up -d
```

Compose сам скачает образ `yamaha226/lms-system:latest` с Docker Hub, поднимет рядом PostgreSQL и накатит миграции.

Проверить, что всё живо:

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

Открыть Swagger в браузере: <http://localhost:8080/swagger/index.html>

Чтобы остановить:

```bash
docker compose down
```

Если хочешь снести и базу — `docker compose down -v`.

## Запуск из исходников (для разработки)

Если хочешь поправить код и запускать локально:

```bash
# Поднять только Postgres в Docker
docker compose up -d postgres

# Накатить миграции
make migrate-up

# Запустить приложение
make run
```

В `.env` поменяй `POSTGRES_HOST=postgres` на `POSTGRES_HOST=localhost`, иначе Go-приложение не найдёт базу.

## API

Все эндпоинты сгруппированы по сущностям: `courses`, `chapters`, `lessons`. Полное описание с примерами запросов и ответов — в Swagger.

| Метод | Путь | Что делает |
|-------|------|------------|
| GET | `/health` | Проверка живости |
| GET | `/courses` | Список всех курсов |
| POST | `/courses` | Создать курс |
| GET | `/courses/:id` | Курс со всеми главами |
| PUT | `/courses/:id` | Обновить курс |
| DELETE | `/courses/:id` | Удалить курс |
| GET | `/courses/:id/chapters` | Главы курса |
| POST | `/chapters` | Создать главу |
| GET | `/chapters/:id` | Главу с уроками |
| PUT | `/chapters/:id` | Обновить главу |
| DELETE | `/chapters/:id` | Удалить главу |
| GET | `/chapters/:id/lessons` | Уроки главы |
| POST | `/lessons` | Создать урок |
| GET | `/lessons/:id` | Урок |
| PUT | `/lessons/:id` | Обновить урок |
| DELETE | `/lessons/:id` | Удалить урок |

### Формат ошибок

Все ошибки приходят в одном виде:

```json
{
  "error": {
    "code": "COURSE_NOT_FOUND",
    "message": "course not found"
  }
}
```

Коды: `INVALID_INPUT` (400), `COURSE_NOT_FOUND` / `CHAPTER_NOT_FOUND` / `LESSON_NOT_FOUND` (404), `INTERNAL_ERROR` (500).

## Конфигурация

Всё читается из переменных окружения (см. `.env.example`):

| Переменная | Что значит | По умолчанию |
|------------|-----------|--------------|
| `APP_PORT` | На каком порту слушает HTTP | `8080` |
| `APP_ENV` | `development` или `production` (влияет на формат логов) | `development` |
| `LOG_LEVEL` | `debug` / `info` / `warn` / `error` | `debug` |
| `POSTGRES_HOST` | Хост базы | `postgres` |
| `POSTGRES_PORT` | Порт базы | `5432` |
| `POSTGRES_USER` | Пользователь | `lms` |
| `POSTGRES_PASSWORD` | Пароль | `lms_password` |
| `POSTGRES_DB` | Имя базы | `lms_db` |
| `POSTGRES_SSLMODE` | SSL для подключения | `disable` |

В `development` логи идут текстом в консоль, в `production` — JSON-ом (удобно собирать в любую систему агрегации).

## Структура проекта

```
cmd/app/                 main, точка входа
internal/
  config/                чтение .env
  entity/                модели БД (GORM)
  repository/            доступ к данным
  service/               бизнес-логика
  dto/                   запросы/ответы API
  handler/               HTTP-обработчики (Gin)
  middleware/            recovery, error handler, request logger
pkg/logger/              обёртка над logrus
migrations/              SQL-миграции (Goose)
docs/                    сгенерированный Swagger
```

Слои общаются через интерфейсы: handler -> service -> repository. Это позволило написать юнит-тесты с моками и не поднимать БД ради проверки логики.

## Миграции

Используется Goose, миграции лежат в `migrations/`. При старте контейнера они накатываются автоматически.

Создать новую миграцию:

```bash
make migrate-create name=add_users_table
```

Накатить / откатить вручную:

```bash
make migrate-up
make migrate-down
```

## Тесты

```bash
make test
```

Покрытие:
- `internal/service` — ~78%
- `internal/handler` — ~76%

Тесты гоняются с флагом `-race`, чтобы ловить race conditions.

Если поправил интерфейс репозитория или сервиса — перегенерь моки:

```bash
make mockery-install   # один раз, поставит CLI
make mocks
```

## Полезные команды

```bash
make help              # список всех таргетов
make build             # собрать бинарь в bin/app
make run               # запустить локально
make test              # тесты
make swag              # перегенерить Swagger-доки
make docker-up         # docker compose up -d
make docker-down       # docker compose down
make migrate-up        # накатить миграции
```

## Docker Hub

Образ: <https://hub.docker.com/r/yamaha226/lms-system>

Теги:
- `latest` — последняя сборка
- `0.1.0` — фиксированная версия

## Что внутри (по этапам стажировки)

1. PostgreSQL поднят через Docker Compose
2. Gin + GORM + Goose-миграции
3. CRUD для Course / Chapter / Lesson через слои repository -> service -> handler
4. Централизованный error handling, единый JSON-формат ошибок
5. Логирование на logrus с уровнями DEBUG / INFO / WARN / ERROR
6. Swagger-документация всех эндпоинтов
7. Юнит-тесты сервисов и хендлеров через Testify + Mockery