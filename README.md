# LMS Main Service

Это бэкенд для системы управления обучением (Learning Management System), который я делал в рамках стажировки в BITLAB ACADEMY.

Сервис умеет хранить курсы, главы и уроки и отдавать их по REST API. Управлять каталогом (создавать/менять/удалять) может только администратор. К урокам можно прикладывать файлы — они лежат в MinIO (S3). Аутентификация и пользователи — через Keycloak (JWT + refresh-токены, роли). Внутри — PostgreSQL, миграции через Goose, документация в Swagger.

## Стек

- Go 1.25
- Gin — HTTP-фреймворк
- GORM — ORM поверх PostgreSQL
- Goose — миграции
- Keycloak — аутентификация, выдача JWT, управление пользователями и ролями
- MinIO — S3-совместимое хранилище для файлов-вложений
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

Compose поднимет приложение, его PostgreSQL, Keycloak с отдельным PostgreSQL, MinIO и одноразовый контейнер, который создаёт bucket для файлов. Образ приложения тянется с Docker Hub, миграции накатываются автоматически, realm Keycloak импортируется при старте, bucket в MinIO создаётся сам — авторизация и загрузка файлов работают из коробки.

Проверить, что всё живо:

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

- Swagger приложения: <http://localhost:8080/swagger/index.html>
- Консоль Keycloak: <http://localhost:8081> (логин `admin` / `admin`)
- Консоль MinIO: <http://localhost:9001> (логин `minioadmin` / `minioadmin`)

> Keycloak стартует ~20 секунд — первые запросы к `/auth/*` могут не пройти, пока он не поднимется.

Чтобы остановить:

```bash
docker compose down
```

Если хочешь снести и базу — `docker compose down -v`.

## Запуск из исходников (для разработки)

Если хочешь поправить код и запускать локально:

```bash
# Поднять зависимости в Docker (всё кроме самого приложения)
docker compose up -d postgres keycloak keycloak-db minio minio-init

# Накатить миграции
make migrate-up

# Запустить приложение
make run
```

В `.env` поменяй `POSTGRES_HOST=postgres` на `POSTGRES_HOST=localhost`, иначе Go-приложение не найдёт базу. `KEYCLOAK_URL` при локальном запуске оставь `http://localhost:8081` (так issuer токенов совпадёт автоматически), а `MINIO_ENDPOINT` — `localhost:9000`.

## API

Полное описание с примерами запросов и ответов — в Swagger. Колонка «Доступ» показывает, что нужно для вызова: 🔓 — публично, 🔒 — валидный JWT, 👑 — JWT с ролью `ROLE_ADMIN`.

### Аутентификация (`/auth`)

| Метод | Путь | Что делает | Доступ |
|-------|------|------------|--------|
| POST | `/auth/login` | Логин по username/password, выдаёт пару токенов | 🔓 |
| POST | `/auth/refresh` | Обновить access-токен по refresh-токену | 🔓 |
| POST | `/auth/register` | Создать пользователя и назначить роль | 👑 |
| PUT | `/auth/profile` | Обновить свои email/имя (роль менять нельзя) | 🔒 |
| PUT | `/auth/password` | Сменить свой пароль (с проверкой текущего) | 🔒 |

### Курсы / главы / уроки

| Метод | Путь | Что делает | Доступ |
|-------|------|------------|--------|
| GET | `/health` | Проверка живости | 🔓 |
| GET | `/courses` | Список всех курсов | 🔓 |
| POST | `/courses` | Создать курс | 👑 |
| GET | `/courses/:id` | Курс со всеми главами | 🔓 |
| PUT | `/courses/:id` | Обновить курс | 👑 |
| DELETE | `/courses/:id` | Удалить курс | 👑 |
| GET | `/courses/:id/chapters` | Главы курса | 🔓 |
| POST | `/chapters` | Создать главу | 👑 |
| GET | `/chapters/:id` | Главу с уроками | 🔓 |
| PUT | `/chapters/:id` | Обновить главу | 👑 |
| DELETE | `/chapters/:id` | Удалить главу | 👑 |
| GET | `/chapters/:id/lessons` | Уроки главы | 🔓 |
| POST | `/lessons` | Создать урок | 👑 |
| GET | `/lessons/:id` | Урок | 🔓 |
| PUT | `/lessons/:id` | Обновить урок | 👑 |
| DELETE | `/lessons/:id` | Удалить урок | 👑 |

Чтение (GET) публично, любые изменения каталога (POST/PUT/DELETE) — только для `ROLE_ADMIN`.

### Файлы-вложения

| Метод | Путь | Что делает | Доступ |
|-------|------|------------|--------|
| POST | `/upload` | Загрузить файл к уроку (`multipart/form-data`: поля `lesson_id` и `file`) | 👑 |
| GET | `/download/:id` | Скачать файл по id вложения | 🔒 |
| GET | `/lessons/:id/attachments` | Список вложений урока (метаданные) | 🔓 |

Сами файлы лежат в MinIO под ключом `lessons/<lesson_id>/<uuid>.<ext>`, а метаданные (имя, размер, content-type, ключ объекта) — в таблице `attachments`. Загружать может только админ, скачивать — любой залогиненный пользователь.

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

| Код | HTTP | Когда |
|-----|------|-------|
| `INVALID_INPUT` | 400 | Не прошла валидация тела/параметров |
| `UNAUTHORIZED` | 401 | Нет токена, токен невалиден или неверный текущий пароль |
| `FORBIDDEN` | 403 | Не хватает роли (например, не админ) |
| `COURSE_NOT_FOUND` / `CHAPTER_NOT_FOUND` / `LESSON_NOT_FOUND` | 404 | Сущность не найдена |
| `CONFLICT` | 409 | Пользователь с таким username/email уже есть |
| `INTERNAL_ERROR` | 500 | Внутренняя ошибка |

## Аутентификация

Пользователи, пароли и роли живут в Keycloak. Приложение проверяет JWT по публичным ключам Keycloak (JWKS) и достаёт роли из токена.

**Тестовые пользователи** (заводятся при импорте realm):

| Логин | Пароль | Роль |
|-------|--------|------|
| `admin` | `admin123` | `ROLE_ADMIN` |
| `teacher` | `teacher123` | `ROLE_TEACHER` |
| `user` | `user123` | `ROLE_USER` |

Access-токен живёт 5 минут, refresh-токен — 168 часов (7 дней).

**Получить токен и дёрнуть защищённый эндпоинт:**

```bash
# 1. Логинимся, забираем access_token
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}' \
  | python3 -c 'import sys,json;print(json.load(sys.stdin)["access_token"])')

# 2. Создаём курс с этим токеном
curl -X POST http://localhost:8080/courses \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Go for beginners","description":"intro"}'
```

**Загрузить и скачать файл к уроку:**

```bash
# Загрузка (только админ): multipart-форма с lesson_id и file
curl -X POST http://localhost:8080/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F lesson_id=1 \
  -F file=@./syllabus.pdf
# {"id":1,"lesson_id":1,"file_name":"syllabus.pdf",...}

# Скачивание (любой залогиненный): по id вложения
curl http://localhost:8080/download/1 \
  -H "Authorization: Bearer $TOKEN" -O -J
```

**Роли:**
- `ROLE_ADMIN` — может всё, включая регистрацию пользователей и назначение ролей.
- `ROLE_TEACHER` / `ROLE_USER` — обычные пользователи; могут менять свой профиль и пароль, но не роли.

Сменить свою роль через API нельзя в принципе — поля роли нет в запросах профиля. Назначить роль может только админ при регистрации.

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
| `KEYCLOAK_URL` | Адрес, по которому приложение **ходит** в Keycloak (должен быть доступен) | `http://localhost:8081` |
| `KEYCLOAK_ISSUER_URL` | URL, зашитый в `iss` токена (для проверки подписи). Пусто = берётся `KEYCLOAK_URL` | (пусто) |
| `KEYCLOAK_REALM` | Realm | `lms` |
| `KEYCLOAK_CLIENT_ID` | Клиент бэкенда | `lms-backend` |
| `KEYCLOAK_CLIENT_SECRET` | Секрет клиента | `lms-backend-secret` |
| `KEYCLOAK_ADMIN_USERNAME` | Админ Keycloak для управления юзерами | `admin` |
| `KEYCLOAK_ADMIN_PASSWORD` | Пароль админа Keycloak | `admin` |
| `MINIO_ENDPOINT` | Адрес S3 API, по которому приложение **ходит** в MinIO (`host:port`, без схемы) | `localhost:9000` |
| `MINIO_ACCESS_KEY` | Access key (он же логин MinIO) | `minioadmin` |
| `MINIO_SECRET_KEY` | Secret key (он же пароль MinIO) | `minioadmin` |
| `MINIO_BUCKET` | Bucket для вложений (создаётся автоматически) | `lms-attachments` |
| `MINIO_USE_SSL` | Ходить в MinIO по HTTPS | `false` |

В `development` логи идут текстом в консоль, в `production` — JSON-ом (удобно собирать в любую систему агрегации).

### Почему два URL для Keycloak

Токен всегда содержит `iss` — адрес, по которому Keycloak «представляется» (issuer). Приложение обязано проверять, что `iss` токена совпадает с ожидаемым. Проблема в том, что *снаружи* (с хоста, из браузера) Keycloak доступен как `localhost:8081`, а *изнутри* Docker-сети приложение ходит к нему как `keycloak:8080` — это разные адреса.

Решение:
- У Keycloak зафиксирован `KC_HOSTNAME=http://localhost:8081` + `hostname-backchannel-dynamic` — issuer в токенах всегда стабильный (`localhost:8081`), но backend может обращаться к нему по внутреннему адресу.
- В приложении `KEYCLOAK_URL` (куда ходить за JWKS/токенами) и `KEYCLOAK_ISSUER_URL` (что проверять в `iss`) разделены. Compose выставляет их автоматически: `keycloak:8080` для запросов и `localhost:8081` для issuer.

При локальном запуске (`make run`) оба адреса совпадают (`localhost:8081`), `KEYCLOAK_ISSUER_URL` можно не задавать.

## Структура проекта

```
cmd/app/                 main, точка входа
internal/
  config/                чтение .env
  entity/                модели БД (GORM)
  repository/            доступ к данным
  service/               бизнес-логика (включая auth_service)
  dto/                   запросы/ответы API
  handler/               HTTP-обработчики (Gin)
  middleware/            recovery, error handler, request logger, JWT-аутентификация
  auth/                  валидация JWT по JWKS, разбор ролей
  keycloak/              клиент Keycloak (логин, refresh, управление юзерами)
  storage/               клиент MinIO (S3): загрузка/скачивание файлов
pkg/logger/              обёртка над logrus
migrations/              SQL-миграции (Goose)
keycloak/import/         realm-export.json — роли и тестовые юзеры
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
- `internal/service` — ~80%
- `internal/handler` — ~82%

Тесты гоняются с флагом `-race`, чтобы ловить race conditions. Keycloak и MinIO в тестах не нужны — клиент Keycloak и хранилище спрятаны за интерфейсами и мокаются через Mockery.

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
- `0.2.0` — фиксированная версия

## Что внутри (по этапам стажировки)

1. PostgreSQL поднят через Docker Compose
2. Gin + GORM + Goose-миграции
3. CRUD для Course / Chapter / Lesson через слои repository -> service -> handler
4. Централизованный error handling, единый JSON-формат ошибок
5. Логирование на logrus с уровнями DEBUG / INFO / WARN / ERROR
6. Swagger-документация всех эндпоинтов
7. Юнит-тесты сервисов и хендлеров через Testify + Mockery
8. Keycloak + отдельный PostgreSQL для него в Docker Compose
9. Интеграция с Keycloak: валидация JWT в middleware, защита мутаций, роли из токена
10. Импорт realm с ролями `ROLE_ADMIN` / `ROLE_TEACHER` / `ROLE_USER` и тестовыми юзерами
11. Эндпоинт логина: выдача access (5 мин) и refresh (168 ч) токенов
12. Эндпоинт обновления access-токена по refresh-токену
13. Регистрация пользователей (только `ROLE_ADMIN`) с назначением роли
14. Обновление профиля и смена пароля (без возможности менять свои роли)
15. CRUD курсов — мутации только для `ROLE_ADMIN`
16. CRUD глав — мутации только для `ROLE_ADMIN`
17. CRUD уроков — мутации только для `ROLE_ADMIN`
18. Интеграция MinIO (S3) через Docker Compose с авто-созданием bucket
19. Эндпоинты `/upload` и `/download` + сущность `Attachment` (файлы к урокам)