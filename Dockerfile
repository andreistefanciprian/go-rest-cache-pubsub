# Build stage
FROM golang:1.25.6-alpine@sha256:bc2596742c7a01aa8c520a075515c7fee21024b05bfaa18bd674fe82c100a05d AS builder

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
FROM gcr.io/distroless/static-debian12:nonroot@sha256:cba10d7abd3e203428e86f5b2d7fd5eb7d8987c387864ae4996cf97191b33764

# Copy the binary from builder stage
COPY --from=builder /app/main /app/main

# Use the nonroot user (UID 65532)
USER nonroot:nonroot

# Expose port
EXPOSE 8080

# Command to run
ENTRYPOINT ["/app/main"]
