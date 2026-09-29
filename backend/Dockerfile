# Stage 1: Build the Go binary
FROM golang:alpine AS builder

WORKDIR /app

# Install dependencies first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

# Stage 2: Minimal runtime image
FROM alpine:latest

WORKDIR /

# Copy the compiled binary from the builder stage
COPY --from=builder /app/server /server

# Expose backend port
EXPOSE 8080

# Command to run the executable
CMD ["/server"]
