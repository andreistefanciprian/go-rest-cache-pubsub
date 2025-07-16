package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// UserCacher defines the interface for caching user data
type UserCacher interface {
	Get(key string) (*User, error)
	Set(key string, user *User) error
	Del(key string) error
}

// RedisUserCacher implements UserCacher interface for Redis cache
type RedisUserCacher struct {
	client *redis.Client
}

// NewRedisUserCacher creates a new Redis client for caching user data
func NewRedisUserCacher(cache *redis.Client) UserCacher {
	return &RedisUserCacher{client: cache}
}

// Del deletes a user from the Redis cache by key
func (r *RedisUserCacher) Del(key string) error {
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
func (r *RedisUserCacher) Get(key string) (*User, error) {
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

// Set stores a user in the Redis cache
// It marshals the user data to JSON format before storing
func (r *RedisUserCacher) Set(key string, user *User) error {
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
