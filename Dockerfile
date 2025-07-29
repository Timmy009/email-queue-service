# Stage 1: Build the Go application
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application
# CGO_ENABLED=0 is important for static binaries in Alpine
# -a for force rebuilding packages that are already up-to-date
# -installsuffix nocgo to avoid issues with CGO
# -ldflags="-s -w" to strip debug information and symbol table for smaller binary
RUN CGO_ENABLED=0 go build -o /bin/email-service ./cmd/email-service

# Stage 2: Create the final minimal image
FROM alpine:latest

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /bin/email-service /usr/local/bin/email-service

# Expose the port the application listens on
EXPOSE 8080

# Run the application
ENTRYPOINT ["/usr/local/bin/email-service"]
