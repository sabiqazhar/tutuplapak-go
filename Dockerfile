# ---- STAGE 1: Build ----
FROM golang:1.24-alpine AS builder

# Install git & ca-certificates (dibutuhkan migrate)
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Install migrate CLI
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Copy seluruh kode
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd

# ---- STAGE 2: Final ----
FROM alpine:latest

# Install CA certificates & bash (untuk script)
RUN apk --no-cache add ca-certificates tzdata bash

# Set working directory
WORKDIR /root/

# Copy binary dari stage builder
COPY --from=builder /app/main .

# Copy migrate binary
COPY --from=builder /go/bin/migrate .

# Copy migrations folder
COPY --from=builder /app/migrations ./migrations/

# Buat entrypoint script
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

# Port yang digunakan aplikasi
EXPOSE 8080

# Jalankan entrypoint
ENTRYPOINT ["./entrypoint.sh"]