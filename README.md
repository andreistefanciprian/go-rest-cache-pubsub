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

## 🛠️ Tech Stack

- **Backend**: Go 1.24.0
- **Database**: PostgreSQL 15
- **Cache**: Redis 7
- **ORM**: GORM
- **Containerization**: Docker & Docker Compose
- **Base Image**: Google Distroless (security-focused)

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose
- Git

### 1. Clone the Repository
```bash
git clone <repository-url>
cd learn_caching
```

### 2. Start the Application
```bash
# Build and start all services
docker-compose up --build

# Or run in background
docker-compose up -d --build
```

### 3. Test the API
```bash
# Create a user
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe"}'

# Get user by ID (first request - cache miss)
curl http://localhost:8080/users/1

# Get user by ID again (cache hit - faster response)
curl http://localhost:8080/users/1
```

## 🔧 Configuration

The application uses environment variables for configuration:

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | Database username |
| `DB_PASSWORD` | `password` | Database password |
| `DB_NAME` | `users` | Database name |
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `REDIS_PASSWORD` | `redispassword` | Redis password |

## 📊 Caching Demonstration

### Cache Miss (First Request)
```bash
curl http://localhost:8080/users/1
```
**Logs show**: `Cache MISS → Database query → Cache update`

### Cache Hit (Subsequent Requests)
```bash
curl http://localhost:8080/users/1
```
**Logs show**: `Cache HIT → Direct response from Redis`

## 🐳 Docker Services

| Service | Port | Description |
|---------|------|-------------|
| `app` | 8080 | Go REST API application |
| `postgres` | 5432 | PostgreSQL database |
| `redis` | 6379 | Redis cache |

## 📝 Example Usage

```bash
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
```

## 🔍 Monitoring

### View Application Logs
```bash
docker-compose logs -f app
```

### View All Services Logs
```bash
docker-compose logs -f
```

### Check Service Health
```bash
docker-compose ps
```

## 🛑 Stopping the Application

```bash
# Stop services
docker-compose down

# Stop and remove volumes (clears data)
docker-compose down -v
```

## 🏁 Learning Outcomes

This project demonstrates:
- **Caching Patterns**: Cache-aside pattern implementation
- **Performance Optimization**: Reduced database load through intelligent caching
- **Microservices Architecture**: Service separation and communication
- **Docker Best Practices**: Multi-stage builds, distroless images, non-root users
- **REST API Design**: Proper HTTP methods and status codes
- **Error Handling**: Graceful degradation and proper error responses

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is for educational purposes to demonstrate caching concepts in Go applications.
