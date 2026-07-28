# ============================
# Stage 1: Build
# ============================
FROM golang:alpine AS builder

WORKDIR /app

# Install dependencies OS yang dibutuhkan untuk build
RUN apk add --no-cache gcc musl-dev

# Copy dependency files dulu (untuk cache layer Docker)
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh source code
COPY . .

# Build binary Go (static, tanpa CGO untuk kompatibilitas maksimal)
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# ============================
# Stage 2: Runtime (minimal)
# ============================
FROM alpine:3.22

WORKDIR /app

# Install ca-certificates untuk HTTPS requests
RUN apk add --no-cache ca-certificates tzdata

# Copy binary dari stage builder
COPY --from=builder /app/server .

# Buat direktori untuk database SQLite
RUN mkdir -p /app/data

# Expose port (sesuaikan dengan env PORT)
EXPOSE 5000

# Jalankan server
CMD ["./server"]
