# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy dependency files first (for layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/main .

# Final stage
FROM alpine:3.19

# Add ca-certificates for HTTPS calls (needed for Cognito)
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./main"]
