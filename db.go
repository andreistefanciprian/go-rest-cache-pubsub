package main

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// UserStorer defines the interface for user storage operations
type UserStorer interface {
	CreateUser(user *User) error
	GetUsers() ([]User, error)
	GetUser(id int) (*User, error)
	UpdateUser(user *User, updates User) error
	DeleteUser(id int) error
}

// User represents a user in the system
// It includes gorm.Model which provides ID, CreatedAt, UpdatedAt, DeletedAt fields
type User struct {
	gorm.Model
	Name string `json:"name"`
}

// PostgreSQLUserStorer implements UserStorer interface for PostgreSQL database
type PostgreSQLUserStorer struct {
	db *gorm.DB
}

// NewPostgreSQLUserStorer creates a new PostgreSQLUserStorer instance
// It initializes the database connection and returns the storer
func NewPostgreSQLUserStorer(db *gorm.DB) UserStorer {
	return &PostgreSQLUserStorer{db: db}
}

// UpdateUser updates an existing user in the database
// It first checks if the user exists, then updates the name field
func (r *PostgreSQLUserStorer) UpdateUser(user *User, updates User) error {
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
func (r *PostgreSQLUserStorer) DeleteUser(id int) error {
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
func (r *PostgreSQLUserStorer) GetUser(id int) (*User, error) {
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
func (r *PostgreSQLUserStorer) CreateUser(user *User) error {
	result := r.db.Create(user)
	if result.Error != nil {
		fmt.Println("Error creating user:", result.Error)
		return result.Error
	}
	fmt.Println("User created:", user)
	return nil
}

// GetUsers retrieves all users from the database
func (r *PostgreSQLUserStorer) GetUsers() ([]User, error) {
	var users []User
	result := r.db.Find(&users)
	if result.Error != nil {
		fmt.Println("Error retrieving users:", result.Error)
		return nil, result.Error
	}
	fmt.Println("Users retrieved:", users)
	return users, nil
}
