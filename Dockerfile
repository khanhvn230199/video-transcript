# Build stage
FROM golang:1.24.0-alpine AS builder

# Cài đặt git và build dependencies
RUN apk add --no-cache git ca-certificates

# Tạo thư mục làm việc
WORKDIR /app

# Copy go mod files để cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Runtime stage
FROM alpine:latest


WORKDIR /app

# Copy binary từ builder stage
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]

