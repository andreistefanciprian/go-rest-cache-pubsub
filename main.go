package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"context"

	"strconv"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(user *User) error
	GetUsers() ([]User, error)
	GetUser(id int) (*User, error)
	UpdateUser(user *User, updates User) error
	DeleteUser(id int) error
}

type userRepo struct {
	db *gorm.DB
}

// User represents a user in the system
// It includes gorm.Model which provides ID, CreatedAt, UpdatedAt, DeletedAt fields
type User struct {
	gorm.Model
	Name string `json:"name"`
}

func (r *userRepo) UpdateUser(user *User, updates User) error {
	// Simulate database delay
	time.Sleep(500 * time.Millisecond)
	fmt.Printf("Updating user %d in database\n", user.ID)
	// First check if user exists
	var existingUser User
	result := r.db.First(&existingUser, user.ID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			fmt.Printf("User %d not found in database\n", user.ID)
			return gorm.ErrRecordNotFound
		}
		fmt.Println("Error checking user existence:", result.Error)
		return result.Error
	}
	// User exists, proceed with update
	existingUser.Name = updates.Name
	result = r.db.Save(&existingUser)
	if result.Error != nil {
		fmt.Println("Error updating user:", result.Error)
		return result.Error
	}
	fmt.Printf("User %d updated in database: %+v\n", user.ID, existingUser)
	return nil
}

// DeleteUser removes a user by ID from the database
func (r *userRepo) DeleteUser(id int) error {
	// Simulate database delay
	time.Sleep(500 * time.Millisecond)
	fmt.Printf("Deleting user %d from database\n", id)

	// First check if user exists
	var user User
	result := r.db.First(&user, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			fmt.Printf("User %d not found in database\n", id)
			return gorm.ErrRecordNotFound
		}
		fmt.Println("Error checking user existence:", result.Error)
		return result.Error
	}

	// User exists, proceed with deletion
	result = r.db.Delete(&user, id)
	if result.Error != nil {
		fmt.Println("Error deleting user:", result.Error)
		return result.Error
	}
	fmt.Printf("User %d deleted from database\n", id)
	return nil
}

// GetUser retrieves a user by ID from the database
func (r *userRepo) GetUser(id int) (*User, error) {
	// Simulate database delay
	time.Sleep(500 * time.Millisecond)
	fmt.Printf("Getting user %d from database\n", id)

	var user User
	result := r.db.First(&user, id)
	if result.Error != nil {
		fmt.Println("Error retrieving user:", result.Error)
		return nil, result.Error
	}
	fmt.Println("User retrieved from database:", user)
	return &user, nil
}

// CreateUser creates a new user in the database
func (r *userRepo) CreateUser(user *User) error {
	result := r.db.Create(user)
	if result.Error != nil {
		fmt.Println("Error creating user:", result.Error)
		return result.Error
	}
	fmt.Println("User created:", user)
	return nil
}

// GetUsers retrieves all users from the database
func (r *userRepo) GetUsers() ([]User, error) {
	var users []User
	result := r.db.Find(&users)
	if result.Error != nil {
		fmt.Println("Error retrieving users:", result.Error)
		return nil, result.Error
	}
	fmt.Println("Users retrieved:", users)
	return users, nil
}

func newUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

type UserCache interface {
	Get(key string) (*User, error)
	Set(key string, user *User) error
	Del(key string) error
}

type redisClient struct {
	client *redis.Client
}

// newRedisClient creates a new Redis client for caching user data
func newRedisClient(cache *redis.Client) UserCache {
	return &redisClient{client: cache}
}

// Del deletes a user from the Redis cache by key
func (r *redisClient) Del(key string) error {
	result, err := r.client.Del(context.Background(), key).Result()
	if err != nil {
		fmt.Println("Error deleting user from cache:", err)
		return err
	}
	if result == 0 {
		fmt.Printf("Cache MISS: User %s not found in Redis\n", key)
		return nil // No error for cache miss, just log it
	}
	// Successfully deleted from cache
	fmt.Println("User deleted from cache:", key)
	return nil
}

// Get retrieves a user from the Redis cache by key
func (r *redisClient) Get(key string) (*User, error) {
	val, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		// Check if it's a cache miss (key not found)
		if err == redis.Nil {
			fmt.Printf("Cache MISS: User %s not found in Redis, checking database\n", key)
			return nil, nil // Return nil user and nil error for cache miss
		}
		// This is an actual Redis error (connection issues, etc.)
		fmt.Println("Error retrieving user from cache:", err)
		return nil, err
	}

	// Key exists in cache, unmarshal the data
	var user User
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		fmt.Println("Error unmarshalling user data:", err)
		return nil, err
	}
	fmt.Printf("Cache HIT: Retrieved user %s from Redis\n", key)
	return &user, nil
}

func (r *redisClient) Set(key string, user *User) error {
	userData, err := json.Marshal(user)
	if err != nil {
		fmt.Println("Error marshalling user data:", err)
		return err
	}
	err = r.client.Set(context.Background(), key, userData, 0).Err()
	if err != nil {
		fmt.Println("Error updating user in cache:", err)
		return err
	}
	fmt.Println("User stored in cache:", *user)
	return nil
}

type UserHandler interface {
	CreateUser(w http.ResponseWriter, r *http.Request)
	GetAllUsers(w http.ResponseWriter, r *http.Request)
	GetUserById(w http.ResponseWriter, r *http.Request)
	UpdateUserById(w http.ResponseWriter, r *http.Request)
	DeleteUserById(w http.ResponseWriter, r *http.Request)
}

type UserService struct {
	repo  UserRepository
	cache UserCache
}

func NewUserService(repo UserRepository, cache UserCache) *UserService {
	return &UserService{
		repo:  repo,
		cache: cache,
	}
}

func validateId(id string) (int, error) {
	if id == "" {
		return 0, fmt.Errorf("ID is required")
	}
	idInt, err := strconv.Atoi(id)
	if err != nil || idInt <= 0 {
		return 0, fmt.Errorf("invalid ID format")
	}
	return idInt, nil
}

// UpdateUserById handles the update of a user by ID
// It expects the ID to be passed in the request path as /users/{id}
// It updates the user in both the database and the cache
// It returns the updated user in JSON format or an error if the update fails
func (s *UserService) UpdateUserById(w http.ResponseWriter, r *http.Request) {
	id, err := validateId(r.PathValue("id"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err.Error()), http.StatusBadRequest)
		fmt.Println("Error:", err.Error())
		return
	}

	var updates User
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Error: Invalid request body", http.StatusBadRequest)
		fmt.Println("Error: Invalid request body")
		return
	}

	if updates.Name == "" {
		http.Error(w, "Error: Name is required", http.StatusBadRequest)
		fmt.Println("Error: Name is required")
		return
	}

	user := &User{Model: gorm.Model{ID: uint(id)}, Name: updates.Name}
	if err := s.repo.UpdateUser(user, updates); err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Error: User not found", http.StatusNotFound)
			fmt.Printf("Error: User %d not found\n", id)
			return
		}
		http.Error(w, "Error: Failed to update user", http.StatusInternalServerError)
		fmt.Println("Error: Failed to update user:", err)
		return
	}

	if err := s.cache.Set(strconv.Itoa(id), user); err != nil {
		http.Error(w, "Error: Failed to update user in cache", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// DeleteUserById handles the deletion of a user by ID
// It deletes the user from both the database and the cache
// It expects the ID to be passed in the request path as /users/{id}
// It returns a 204 No Content status if successful, or an error if not
func (s *UserService) DeleteUserById(w http.ResponseWriter, r *http.Request) {
	id, err := validateId(r.PathValue("id"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err.Error()), http.StatusBadRequest)
		fmt.Println("Error:", err.Error())
		return
	}

	// Try to delete from database first
	if err := s.repo.DeleteUser(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Error: User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error: Failed to delete user", http.StatusInternalServerError)
		fmt.Println("Error: Failed to delete user:", err)
		return
	}

	// Delete from cache
	if err := s.cache.Del(strconv.Itoa(id)); err != nil {
		http.Error(w, "Error: Failed to delete user from cache", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetUserById handles the retrieval of a user by ID
// It first checks the cache, and if not found, retrieves from the database
// It expects the ID to be passed in the request path as /users/{id}
// It returns the user in JSON format or an error if not found
func (s *UserService) GetUserById(w http.ResponseWriter, r *http.Request) {
	id, err := validateId(r.PathValue("id"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err.Error()), http.StatusBadRequest)
		fmt.Println("Error:", err.Error())
		return
	}
	// Check cache first
	user, err := s.cache.Get(strconv.Itoa(id))
	if err != nil {
		http.Error(w, "Error: Failed to retrieve user from cache", http.StatusInternalServerError)
		return
	}
	if user != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
		return
	}
	// If not found in cache, retrieve from database
	user, err = s.repo.GetUser(id)
	if err == gorm.ErrRecordNotFound {
		http.Error(w, "Error: User not found", http.StatusNotFound)
		fmt.Println("Error: User not found")
		return
	}
	if err != nil {
		http.Error(w, "Error: Failed to retrieve user", http.StatusInternalServerError)
		fmt.Println("Error: Failed to retrieve user")
		return
	}
	if user == nil {
		http.Error(w, "Error: User not found", http.StatusNotFound)
		fmt.Println("Error: User not found")
		return
	}
	// Store in cache for future requests
	if err := s.cache.Set(strconv.Itoa(id), user); err != nil {
		http.Error(w, "Error: Failed to store user in cache", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// CreateUser handles the creation of a new user
// It expects the user data in the request body as JSON
// It validates the input and returns the created user in JSON format or an error if creation fails
func (s *UserService) CreateUser(w http.ResponseWriter, r *http.Request) {
	user := User{}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Error: Invalid request body", http.StatusBadRequest)
		fmt.Println("Error: Invalid request body")
		return
	}
	if user.Name == "" {
		http.Error(w, "Error: Name is required", http.StatusBadRequest)
		fmt.Println("Error: Name is required")
		return
	}
	if err := s.repo.CreateUser(&user); err != nil {
		http.Error(w, "Error: Failed to create user", http.StatusInternalServerError)
		fmt.Println("Error: Failed to create user")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// GetAllUsers handles the retrieval of all users
// It retrieves users from the database and returns them in JSON format
// If no users are found, it returns a 204 No Content status
// If an error occurs, it returns a 500 Internal Server Error status
func (s *UserService) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.repo.GetUsers()
	if err != nil {
		http.Error(w, "Error: Failed to retrieve users", http.StatusInternalServerError)
		fmt.Println("Error: Failed to retrieve users")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if len(users) == 0 {
		w.WriteHeader(http.StatusNoContent)
		fmt.Println("No users found")
		return
	}
	json.NewEncoder(w).Encode(users)
}

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
	userRepository := newUserRepository(db)

	// Initialize Redis cache
	cache, err := initCache()
	if err != nil {
		fmt.Println("Error initializing Redis cache:", err)
		return
	}
	userCache := newRedisClient(cache)

	handler := NewUserService(userRepository, userCache)

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
