# Build stage
FROM golang:1.21-alpine AS builder

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

# Cài đặt FFmpeg và ca-certificates
RUN apk add --no-cache ffmpeg ca-certificates

# Tạo user không phải root để chạy app
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binary từ builder stage
COPY --from=builder /app/main .

# Copy templates nếu có
COPY --from=builder /app/templates ./templates

# Tạo thư mục uploads và set permissions
RUN mkdir -p uploads && \
    chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]

