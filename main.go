package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name string `json:"name"`
}

type Users struct {
	cache *redis.Client
	db    *gorm.DB
}

func (u *Users) createUser(w http.ResponseWriter, r *http.Request) {
	user := User{}

	json.NewDecoder(r.Body).Decode(&user)
	if user.Name == "" {
		http.Error(w, "Error: Name is required", http.StatusBadRequest)
		fmt.Println("Error: Name is required")
		return
	}
	u.db.Create(&user)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
	fmt.Println("User created:", user)
}

func (u *Users) getAllUsers(w http.ResponseWriter, r *http.Request) {

	users := []User{}
	u.db.Find(&users)
	if len(users) == 0 {
		http.Error(w, "Error: No users found", http.StatusNotFound)
		fmt.Println("Error: No users found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
	fmt.Println("All users retrieved")
}

func (u *Users) updateUserById(w http.ResponseWriter, r *http.Request) {
	newUser := User{}
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, "Error: Invalid request body", http.StatusBadRequest)
		fmt.Println("Error: Invalid request body")
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Error: Invalid User Id", http.StatusBadRequest)
		fmt.Println("Error: Invalid user ID")
		return
	}
	if newUser.Name == "" {
		http.Error(w, "Error: Name is required", http.StatusBadRequest)
		fmt.Println("Error: Name is required")
		return
	}
	if id <= 0 {
		http.Error(w, "Error: User ID must be greater than 0", http.StatusBadRequest)
		fmt.Println("Error: User ID must be greater than 0")
		return
	}
	// Check if user exists in database
	currentUser, err := u.retrieveUserFromDb(id)
	if err != nil {
		http.Error(w, "Error: User not found", http.StatusNotFound)
		fmt.Printf("Error: User %d not found\n", id)
		return
	}
	// Update user in database
	if err := u.updateUserInDB(currentUser, User{Name: newUser.Name}); err != nil {
		http.Error(w, "Error: Failed to update User", http.StatusInternalServerError)
		fmt.Println("Error: Failed to update user")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentUser)
	// Update user in cache
	userData, _ := json.Marshal(currentUser)
	err = u.cache.Set(context.Background(), strconv.Itoa(id), userData, 0).Err()
	if err != nil {
		fmt.Println("Error storing user in cache:", err)
	} else {
		fmt.Println("User updated in cache:", currentUser)
	}

}

func (u *Users) deleteUserById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		http.Error(w, "Error: Invalid User Id", http.StatusBadRequest)
		fmt.Println("Error: Invalid user ID")
		return
	}

	user, err := u.retrieveUserFromDb(id)
	if err != nil {
		http.Error(w, "Error: User not found", http.StatusNotFound)
		fmt.Printf("Error: User %d not found\n", id)
		return
	}
	u.db.Delete(&user)
	fmt.Println("User deleted:", user)
	// Remove user from cache
	err = u.cache.Del(context.Background(), strconv.Itoa(id)).Err()
	if err != nil {
		fmt.Println("Error removing user from cache:", err)
	} else {
		fmt.Println("User removed from cache:", id)
	}
	json.NewEncoder(w).Encode(user)
}

func (u *Users) getUserById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		http.Error(w, "Error: Invalid User ID", http.StatusBadRequest)
		fmt.Println("Error: Invalid user ID")
		return
	}
	// Check if user exists in cache
	user, err := u.retrieveUserFromCache(id)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
		return
	}
	// If not found in cache, check in db
	user, err = u.retrieveUserFromDb(id)
	if err != nil {
		http.Error(w, "Error: User not found", http.StatusNotFound)
		fmt.Printf("Error: User %d not found\n", id)
		return
	}
	// If found in db, return user
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
	// Store user in cache
	if err := u.updateUserInCache(id, user); err != nil {
		http.Error(w, "Error: Failed to update user in cache", http.StatusInternalServerError)
		fmt.Println("Error: Failed to update user in cache")
		return
	}
}

func (u *Users) retrieveUserFromCache(id int) (*User, error) {
	val, err := u.cache.Get(context.Background(), strconv.Itoa(id)).Result()
	if err != nil {
		fmt.Println("Error retrieving user from cache:", err)
		return nil, err
	}
	var user User
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		fmt.Println("Error unmarshalling user data:", err)
		return nil, err
	}
	fmt.Println("User retrieved from cache:", user)
	return &user, nil
}

func (u *Users) updateUserInCache(id int, user *User) error {
	userData, err := json.Marshal(user)
	if err != nil {
		fmt.Println("Error marshalling user data:", err)
		return err
	}
	err = u.cache.Set(context.Background(), strconv.Itoa(id), userData, 0).Err()
	if err != nil {
		fmt.Println("Error updating user in cache:", err)
		return err
	}
	fmt.Println("User stored in cache:", user)
	return nil
}

// updateUserInDB updates a user in the database
func (u *Users) updateUserInDB(user *User, updates User) error {
	result := u.db.Model(user).Updates(updates)
	if result.Error != nil {
		fmt.Println("Error updating user in DB:", result.Error)

		return result.Error
	}
	fmt.Println("User updated in db:", user)
	return nil
}

func (u *Users) retrieveUserFromDb(id int) (*User, error) {
	var user User
	result := u.db.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	fmt.Println("User retrieved from db:", user)
	return &user, nil
}

// initDbAndCache initializes the database and cache connections
func initDbAndCache() (*Users, error) {
	users := &Users{}

	// Get configuration from environment variables with defaults
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisPassword := getEnv("REDIS_PASSWORD", "redispassword")

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "users")

	// Initialize Redis client
	users.cache = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPassword,
		DB:       0,
	})
	ctx := context.Background()
	pong, err := users.cache.Ping(ctx).Result()
	fmt.Println("Redis ping:", pong, err)

	// Initialize database connection
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		dbHost, dbUser, dbPassword, dbName, dbPort)
	users.db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate the User model
	if err := users.db.AutoMigrate(&User{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return users, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// Initialize database and cache
	users, err := initDbAndCache()
	if err != nil {
		fmt.Println("Error initializing database and cache:", err)
		return
	}
	// Set up HTTP server and routes
	fmt.Println("Server is starting on port 8080...")
	mux := http.NewServeMux()
	mux.HandleFunc("POST /users", users.createUser)
	mux.HandleFunc("GET /users", users.getAllUsers)
	mux.HandleFunc("GET /users/{id}", users.getUserById)
	mux.HandleFunc("PUT /users/{id}", users.updateUserById)
	mux.HandleFunc("DELETE /users/{id}", users.deleteUserById)
	http.ListenAndServe(":8080", mux)
}
