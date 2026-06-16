# Stage 1: Build the Go binary
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Download dependencies first (cached if go.mod/go.sum unchanged)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./main.go

# Stage 2: Minimal runtime image
FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
