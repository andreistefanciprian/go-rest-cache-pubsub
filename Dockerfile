# Build stage
FROM golang:1.25.2-alpine@sha256:182059d7dae0e1dfe222037d14b586ebece3ebf9a873a0fe1cc32e53dbea04e0 AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage - using distroless
FROM gcr.io/distroless/static-debian12:nonroot@sha256:e8a4044e0b4ae4257efa45fc026c0bc30ad320d43bd4c1a7d5271bd241e386d0

# Copy the binary from builder stage
COPY --from=builder /app/main /app/main

# Use the nonroot user (UID 65532)
USER nonroot:nonroot

# Expose port
EXPOSE 8080

# Command to run
ENTRYPOINT ["/app/main"]
