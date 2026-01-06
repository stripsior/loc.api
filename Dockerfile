# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
# gcc, musl-dev are needed for CGO
# oniguruma-dev is needed for go-enry regex support
RUN apk add --no-cache gcc musl-dev git oniguruma-dev

WORKDIR /app

# Copy go mod and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# CGO_ENABLED=1 is required for oniguruma support in go-enry
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o loc-api ./cmd/server/main.go

# Final stage
FROM alpine:3.20

# Install runtime dependencies
# ca-certificates for HTTPS communication with GitHub
# git is used for cloning repositories
# oniguruma is needed for go-enry at runtime
RUN apk add --no-cache ca-certificates git oniguruma

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/loc-api .

# Expose the API port
EXPOSE 8080

# Environment variables with defaults
ENV PORT=8080
ENV CACHE_TTL_MINUTES=60
ENV CORS_ALLOW_ORIGIN=*

# Run the application
CMD ["./loc-api"]
