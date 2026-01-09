FROM golang:1.24-alpine AS builder

RUN apk add --no-cache build-base

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server .

COPY static ./static

COPY db/migrations ./migrations

EXPOSE 8080

# Run binary
CMD ["./server"]

# docker build -t gasthaus-backend .
# docker run -p 8080:8080 --env-file .env gasthaus-backend
