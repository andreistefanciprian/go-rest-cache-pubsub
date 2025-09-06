# Learn Caching - Go REST API with Redis Cache

A Go-based REST API application demonstrating caching patterns using Redis and PostgreSQL. This project showcases how to implement efficient data retrieval with cache-first strategy to improve application performance.

## 🚀 Features

- **REST API**: Full CRUD operations for user management
- **Redis Caching**: Cache-first strategy with automatic fallback to database
- **PostgreSQL Database**: Persistent data storage with GORM ORM
- **Docker Support**: Fully containerized application with Docker Compose
- **Security**: Distroless Docker images with non-root user
- **Environment Configuration**: Flexible configuration via environment variables

## 🏗️ Architecture

```
HTTP Request → Cache Check (Redis) → Database (PostgreSQL) → Cache Update
```

### Caching Strategy:
1. **Cache Hit**: Data retrieved directly from Redis (fast response)
2. **Cache Miss**: Data fetched from PostgreSQL and stored in Redis for future requests
3. **Cache Invalidation**: Cache updated on data modifications (UPDATE/DELETE)

## 📋 API Endpoints

| Method | Endpoint | Description | Caching |
|--------|----------|-------------|---------|
| `POST` | `/users` | Create a new user | ❌ |
| `GET` | `/users` | Get all users | ❌ |
| `GET` | `/users/{id}` | Get user by ID | ✅ Cache-first |
| `PUT` | `/users/{id}` | Update user by ID | ✅ Cache update |
| `DELETE` | `/users/{id}` | Delete user by ID | ✅ Cache invalidation |

## 🚀 Run it

```
# Build and start all services
docker-compose up --build

# Create users
curl -X POST http://localhost:8080/users -H "Content-Type: application/json" -d '{"name": "Alice"}'
curl -X POST http://localhost:8080/users -H "Content-Type: application/json" -d '{"name": "Bob"}'

# Get all users
curl http://localhost:8080/users

# Get specific user (demonstrates caching)
curl http://localhost:8080/users/1  # Cache miss - slower
curl http://localhost:8080/users/1  # Cache hit - faster

# Update user (updates cache)
curl -X PUT http://localhost:8080/users/1 -H "Content-Type: application/json" -d '{"name": "Alice Updated"}'

# Delete user (removes from cache)
curl -X DELETE http://localhost:8080/users/1

# View Application Logs
docker-compose logs -f app

# Check Service Health
docker-compose ps

# Stop services
docker-compose down

# Stop and remove volumes (clears data)
docker-compose down -v
```

## 🛠️ Tech Stack

- **Backend**: Go 1.24.0
- **Database**: PostgreSQL 15
- **Cache**: Redis 7
- **ORM**: GORM
- **Containerization**: Docker & Docker Compose
- **Base Image**: Google Distroless (security-focused)