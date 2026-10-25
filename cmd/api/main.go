package main

import (
	"advanced-blog-management-system/internal/apperrors"
	"advanced-blog-management-system/internal/eventlog"
	"advanced-blog-management-system/internal/handler"
	"advanced-blog-management-system/internal/middleware"
	"advanced-blog-management-system/internal/repository"
	"advanced-blog-management-system/internal/service"
	"advanced-blog-management-system/pkg/database"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Загрузка .env
	if err := godotenv.Load(); err != nil {
		log.Println("warning: .env file not found, using environment variables")
	}
	log.Println("configuration loading...")

	// 2. Конфигурация
	dbCfg := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "blog_db"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	serverHost := getEnv("SERVER_HOST", "0.0.0.0")
	serverPort := getEnvInt("SERVER_PORT", 8080)

	// 3. Подключение к БД
	log.Println("connecting to database...")
	db, err := database.NewPostgresDB(dbCfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// 4. Миграции
	log.Println("running migrations...")
	migrations := []string{
		mustReadFile("migrations/001_init_schema.sql"),
		mustReadFile("migrations/002_add_indexes.sql"),
	}
	if err := database.RunMigrations(db, migrations); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	// 5. Репозитории
	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	commentRepo := repository.NewCommentRepository(db)

	// 6. Сервисы
	userService := service.NewUserService(userRepo)
	postService := service.NewPostService(postRepo, userRepo, commentRepo)
	commentService := service.NewCommentService(commentRepo, postRepo, userRepo)

	// 7. Асинхронный логгер событий (канал + горутина + log.txt)
	eventLogger, err := eventlog.New("log.txt", 1500*time.Millisecond)
	if err != nil {
		log.Fatalf("failed to init event logger: %v", err)
	}
	// Закроем в конце main (после shutdown сервера), с таймаутом.
	// ВАЖНО: defer сработает ПОСЛЕ srv.Shutdown, потому что main дойдёт до конца.
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		eventLogger.Close(ctx)
	}()

	// 8. Хендлеры
	authHandler := handler.NewAuthHandler(userService, jwtSecret)
	postHandler := handler.NewPostHandler(postService, eventLogger)
	commentHandler := handler.NewCommentHandler(commentService, eventLogger)

	// 9. Роутер
	r := setupRouter(authHandler, postHandler, commentHandler, jwtSecret)

	// 10. HTTP-сервер
	addr := fmt.Sprintf("%s:%d", serverHost, serverPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 11. Запуск в горутине
	go func() {
		log.Printf("server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// 12. Ожидание сигнала
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutdown signal received")

	// 13. Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server forced shutdown: %v", err)
	}
	log.Println("server stopped")
}

func setupRouter(
	authHandler *handler.AuthHandler,
	postHandler *handler.PostHandler,
	commentHandler *handler.CommentHandler,
	jwtSecret string,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.RecoveryMiddleware)
	r.Use(middleware.CORSMiddleware)

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", handler.HealthCheckHandler)
		r.Post("/register", authHandler.RegisterHandler)
		r.Post("/login", authHandler.LoginHandler)

		// Публичные
		r.Get("/posts", postHandler.GetAllPosts)
		r.Get("/posts/{id}", postHandler.GetPost)
		r.Get("/users/{id}/posts", postHandler.GetPostsByAuthor)
		r.Get("/posts/{postId}/comments", commentHandler.GetCommentsByPostID)

		// Защищенные
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(jwtSecret))
			r.Post("/posts", postHandler.CreatePost)
			r.Put("/posts/{id}", postHandler.UpdatePost)
			r.Delete("/posts/{id}", postHandler.DeletePost)
			r.Post("/posts/{postId}/comments", commentHandler.CreateComment)
			r.Put("/comments/{id}", commentHandler.UpdateComment)
			r.Delete("/comments/{id}", commentHandler.DeleteComment)
		})
	})

	return r
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func mustReadFile(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read %s: %v", path, err)
	}
	return string(b)
}

var _ = apperrors.ErrUnauthorized // избегаем unused-import, если не используется
