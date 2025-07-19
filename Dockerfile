# Stage 1: Build the application
FROM golang:1.24-bookworm AS builder

WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "-s -w" -o /app/server ./cmd/api/main.go

# Stage 2: Create the final, minimal image
FROM gcr.io/distroless/base-debian12

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/server .

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["/app/server"]
