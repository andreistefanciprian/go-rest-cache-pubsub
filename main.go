package main

import (
	"fmt"
	"net/http"
	"os"

	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// initDb initializes the database connection and migrates the User model
func initDb() (*gorm.DB, error) {

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "users")

	// Initialize database connection
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		dbHost, dbUser, dbPassword, dbName, dbPort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate the User model
	if err := db.AutoMigrate(&User{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// initCache initializes the Redis cache connection
func initCache() (*redis.Client, error) {

	// Get configuration from environment variables with defaults
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisPassword := getEnv("REDIS_PASSWORD", "redispassword")

	// Initialize Redis client
	cache := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPassword,
		DB:       0,
	})
	ctx := context.Background()
	pong, err := cache.Ping(ctx).Result()
	fmt.Println("Redis ping:", pong, err)

	return cache, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// withNotFoundHandler wraps the HTTP handler to return a 404 Not Found error
// if the requested route does not match any registered handlers
func withNotFoundHandler(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pattern := mux.Handler(r)
		if pattern == "" {
			http.Error(w, "Error: Route not found", http.StatusNotFound)
			fmt.Println("Error: Route not found for", r.Method, r.URL.Path)
			return
		}
		mux.ServeHTTP(w, r)
	})
}

func main() {
	// Initialize database and cache
	db, err := initDb()
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}
	userStorer := NewPostgreSQLUserStorer(db)

	// Initialize Redis cache
	cache, err := initCache()
	if err != nil {
		fmt.Println("Error initializing Redis cache:", err)
		return
	}
	userCacher := NewRedisUserCacher(cache)

	handler := NewUserService(userStorer, userCacher)

	// Set up HTTP server and routes
	fmt.Println("Server is starting on port 8080...")
	mux := http.NewServeMux()
	mux.HandleFunc("POST /users", handler.CreateUser)
	mux.HandleFunc("GET /users", handler.GetAllUsers)
	mux.HandleFunc("GET /users/{id}", handler.GetUserById)
	mux.HandleFunc("PUT /users/{id}", handler.UpdateUserById)
	mux.HandleFunc("DELETE /users/{id}", handler.DeleteUserById)
	http.ListenAndServe(":8080", withNotFoundHandler(mux))
}
