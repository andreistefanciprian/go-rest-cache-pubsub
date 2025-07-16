package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"gorm.io/gorm"
)

// HTTPUserHandler defines the interface for handling user-related HTTP requests
type HTTPUserHandler interface {
	CreateUser(w http.ResponseWriter, r *http.Request)
	GetAllUsers(w http.ResponseWriter, r *http.Request)
	GetUserById(w http.ResponseWriter, r *http.Request)
	UpdateUserById(w http.ResponseWriter, r *http.Request)
	DeleteUserById(w http.ResponseWriter, r *http.Request)
}

// UserService implements HTTPUserHandler interface
// It provides methods to handle user-related HTTP requests
// It uses UserStorer for database operations and UserCacher for caching
type UserService struct {
	storer UserStorer
	cacher UserCacher
}

// NewUserService creates a new UserService instance
// It initializes the service with the provided UserStorer and UserCacher
func NewUserService(storer UserStorer, cacher UserCacher) HTTPUserHandler {
	return &UserService{
		storer: storer,
		cacher: cacher,
	}
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
	if err := s.storer.UpdateUser(user, updates); err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Error: User not found", http.StatusNotFound)
			fmt.Printf("Error: User %d not found\n", id)
			return
		}
		http.Error(w, "Error: Failed to update user", http.StatusInternalServerError)
		fmt.Println("Error: Failed to update user:", err)
		return
	}

	if err := s.cacher.Set(strconv.Itoa(id), user); err != nil {
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
	if err := s.storer.DeleteUser(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Error: User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error: Failed to delete user", http.StatusInternalServerError)
		fmt.Println("Error: Failed to delete user:", err)
		return
	}

	// Delete from cache
	if err := s.cacher.Del(strconv.Itoa(id)); err != nil {
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
	user, err := s.cacher.Get(strconv.Itoa(id))
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
	user, err = s.storer.GetUser(id)
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
	if err := s.cacher.Set(strconv.Itoa(id), user); err != nil {
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
	if err := s.storer.CreateUser(&user); err != nil {
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
	users, err := s.storer.GetUsers()
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

// validateId checks if the provided ID is valid
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
