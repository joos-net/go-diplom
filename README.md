# Advanced Blog Management System
## Островский Евгений

Дипломный проект на Go для разработки полнофункциональной системы управления блогом с REST API, JWT аутентификацией и PostgreSQL.

## 📋 Содержание

- [О проекте](#о-проекте)
- [Структура проекта](#структура-проекта)
- [Технологический стек](#технологический-стек)
- [Быстрый старт](#быстрый-старт)
- [Разработка](#разработка)
- [API эндпоинты](#api-эндпоинты)
- [Тестирование](#тестирование)

## О проекте

**Advanced Blog Management System** - это дипломный проект для студентов, которые изучают Go, REST API и работу с базами данных. 

Проект включает:
- ✅ Управление пользователями (регистрация, вход)
- ✅ Управление постами (создание, редактирование, удаление)
- ✅ Управление комментариями (создание, редактирование, удаление)
- ✅ JWT аутентификация и авторизация
- ✅ Отложенная публикация постов (планировщик)
- ✅ Логирование и обработка ошибок
- ✅ Docker контейнеризация

## Структура проекта

```
template_project/
├── cmd/api/
│   └── main.go                 # Точка входа приложения
├── internal/
│   ├── handler/                # HTTP обработчики
│   │   ├── auth_handler.go
│   │   ├── post_handler.go
│   │   ├── comment_handler.go
│   │   └── health.go
│   ├── middleware/             # HTTP middleware
│   │   ├── auth.go
│   │   └── logging.go
│   ├── model/                  # Модели данных
│   │   └── models.go
│   ├── repository/             # Слой доступа к БД
│   │   ├── interfaces.go
│   │   ├── user_repo.go
│   │   ├── post_repo.go
│   │   └── comment_repo.go
│   └── service/                # Бизнес-логика
│   │   ├── user_service.go
│   │   ├── post_service.go
│   │   └── comment_service.go
│   └── eventlog/
│   │   └── logger.go           # Логирование
│   └── apperrors/
│       └── errors.go           # Ошибки
├── pkg/
│   ├── auth/                   # Утилиты аутентификации
│   │   ├── jwt.go
│   │   └── password.go
│   └── database/               # Утилиты БД
│       └── db.go
├── migrations/                 # SQL миграции
│   ├── 001_init_schema.sql
│   └── 002_add_indexes.sql
├── .env.example                # Пример конфигурации
├── docker-compose.yml          # Docker Compose
├── Dockerfile                  # Docker образ
└── go.mod                      # Зависимости проекта
```

## Технологический стек

- **Язык:** Go 1.24.5
- **Веб-фреймворк:** Chi Router
- **База данных:** PostgreSQL 15
- **Аутентификация:** JWT (golang-jwt)
- **Хеширование:** bcrypt (golang.org/x/crypto)
- **Контейнеризация:** Docker, Docker Compose
- **Валидация:** go-playground/validator

## Быстрый старт

### Предварительные требования

- Go 1.24.5 или выше
- Docker и Docker Compose
- Bash или другая оболочка команд

### 1. Подготовка окружения

```bash
# Клонировать репозиторий
git clone https://github.com/joos-net/go-diplom.git
cd go-diplom

# Скопировать конфигурацию
cp .env.example .env

# Установить Go зависимости
go mod download
```

### 2. Запуск БД

```bash
# Запустить PostgreSQL в Docker
docker-compose up -d db

# Подождать пока БД запустится (примерно 15 секунд)
docker-compose logs db
```

### 3. Запуск приложения

```bash
# Запустить приложение
go run cmd/api/main.go

# Или собрать и запустить
go build -o api ./cmd/api/main.go
./api
```

Приложение будет доступно на `http://localhost:8080`

### 4. Docker контейнеризация

```bash
# Запустить всё через Docker Compose
docker-compose up --build

# Остановить
docker-compose down

# Очистить данные БД
docker-compose down -v
```

## API эндпоинты

### Публичные эндпоинты

```
GET    /api/health                     # Проверка здоровья приложения
POST   /api/register                   # Регистрация нового пользователя
POST   /api/login                      # Вход пользователя
GET    /api/posts                      # Получить все посты
GET    /api/posts/{id}                 # Получить пост по ID
GET    /api/users/{id}/posts           # Получить посты пользователя
GET    /api/posts/{postId}/comments    # Получить комментарии к посту
```

### Защищенные эндпоинты (требуют Authorization: Bearer <token>)

```
POST   /api/posts                      # Создать пост
PUT    /api/posts/{id}                 # Обновить пост
DELETE /api/posts/{id}                 # Удалить пост
POST   /api/posts/{postId}/comments    # Добавить комментарий
PUT    /api/comments/{id}              # Обновить комментарий
DELETE /api/comments/{id}              # Удалить комментарий
```

## Тестирование

### Примеры curl команд

```bash
# Health check
curl http://localhost:8080/api/health

# Регистрация
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'

# Вход и получение токена
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'

# Создание поста (задайте переменную TOKEN)
curl -X POST http://localhost:8080/api/posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "My First Post",
    "content": "This is my first post"
  }'

# Получение всех постов
curl http://localhost:8080/api/posts

# Получение конкретного поста
curl http://localhost:8080/api/posts/1
```

## Конфигурация

Переменные окружения задаются в файле `.env`:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=blog_db
DB_SSLMODE=disable

# JWT
JWT_SECRET=your-secret-key-change-in-production

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Environment
ENV=development
```

## Архитектура приложения

Проект использует чистую архитектуру с разделением ответственности:

```
┌─────────────────┐
│  HTTP Requests  │
└────────┬────────┘
         │
┌────────▼────────────────────────┐
│ Middleware (Auth, Logging, CORS)│
└────────┬────────────────────────┘
         │
┌────────▼─────────────┐
│ Handlers (HTTP API)  │ ← Парсинг JSON, валидация, HTTP коды
└────────┬─────────────┘
         │
┌────────▼──────────────┐
│ Services (Business)   │ ← Бизнес-логика, валидация, права
└────────┬──────────────┘
         │
┌────────▼─────────────┐
│ Repositories (Data)  │ ← SQL запросы, работа с БД
└────────┬─────────────┘
         │
┌────────▼────────┐
│  PostgreSQL DB  │
└─────────────────┘
```

## Базовые концепции

### Context

Все методы БД используют `context.Context` для управления таймаутами и отменой операций:

```go
user, err := repo.GetByID(ctx, userID)
```

### Параметризованные запросы

Для защиты от SQL injection используйте параметризованные запросы:

```go
// ❌ Неправильно
query := fmt.Sprintf("SELECT * FROM users WHERE id = %d", id)

// ✅ Правильно
query := "SELECT * FROM users WHERE id = $1"
db.QueryRow(query, id)
```

### Обработка ошибок

Всегда проверяйте ошибки:

```go
if err != nil {
    if err == sql.ErrNoRows {
        return nil, nil  // Не найдено - это не ошибка
    }
    return nil, fmt.Errorf("failed to query: %w", err)
}
```

### JWT токены

Токены генерируются при регистрации/входе и требуются для защищенных операций:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## Полезные команды

```bash
# Скачать зависимости
go mod download
go mod tidy

# Запустить приложение
go run cmd/api/main.go

# Собрать приложение
go build -o api ./cmd/api/main.go

# Просмотреть логи БД
docker-compose logs db -f

# Подключиться к БД
docker-compose exec db psql -U postgres -d blog_db

# Остановить все сервисы
docker-compose down

# Очистить данные БД
docker-compose down -v
```

## SQL запросы для ручного тестирования

```sql
-- Подключиться к БД
docker-compose exec db psql -U postgres -d blog_db

-- Просмотреть таблицы
\dt

-- Просмотреть пользователей
SELECT id, username, email, created_at FROM users;

-- Просмотреть посты
SELECT id, title, status, author_id, created_at FROM posts;

-- Просмотреть комментарии
SELECT id, content, post_id, author_id, created_at FROM comments;
```
