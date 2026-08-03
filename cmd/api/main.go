package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"platform/internal/handler"
	"platform/internal/middleware"
	"platform/internal/repository"
	"platform/internal/service"
	"platform/pkg/auth"
	"platform/pkg/database"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	commentRepo := repository.NewCommentRepository(db)

	userService := service.NewUserService(userRepo)
	postService := service.NewPostService(postRepo)
	commentService := service.NewCommentService(commentRepo, postRepo)

	jwtManager := auth.NewJWTManager(os.Getenv("JWT_SECRET"))

	authHandler := handler.NewAuthHandler(userService, jwtManager)
	postHandler := handler.NewPostHandler(postService)
	commentHandler := handler.NewCommentHandler(commentService)
	healthHandler := handler.NewHealthHandler(db)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", healthHandler.Health)
	mux.HandleFunc("/api/register", authHandler.Register)
	mux.HandleFunc("/api/login", authHandler.Login)

	mux.HandleFunc("/api/posts", func(w http.ResponseWriter, r *http.Request) {
		authMW := middleware.NewAuthMiddleware(jwtManager)
		switch r.Method {
		case http.MethodGet:
			postHandler.GetAll(w, r)
		case http.MethodPost:
			authMW(http.HandlerFunc(postHandler.Create)).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/posts/", func(w http.ResponseWriter, r *http.Request) {
		authMW := middleware.NewAuthMiddleware(jwtManager)
		path := strings.TrimPrefix(r.URL.Path, "/api/posts/")
		parts := strings.Split(path, "/")

		if len(parts) > 1 && parts[1] == "comments" {
			if err != nil {
				http.Error(w, "Invalid post ID", http.StatusBadRequest)
				return
			}
			switch r.Method {
			case http.MethodGet:
				commentHandler.GetByPostID(w, r)
			case http.MethodPost:
				authMW(http.HandlerFunc(commentHandler.Create)).ServeHTTP(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		switch r.Method {
		case http.MethodGet:
			postHandler.GetByID(w, r)
		case http.MethodPut:
			authMW(http.HandlerFunc(postHandler.Update)).ServeHTTP(w, r)
		case http.MethodDelete:
			authMW(http.HandlerFunc(postHandler.Delete)).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	loggingMiddleware := middleware.LoggingMiddleware
	recoveryMiddleware := middleware.RecoveryMiddleware

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      recoveryMiddleware(loggingMiddleware(mux)),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server stopped")
}
