# Stage 1: Build
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build only the receiver component
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /receiver-app ./cmd/receiver

# Stage 2: Final Image
FROM alpine:3.18
EXPOSE 8080
COPY --from=builder /receiver-app /receiver-app
ENTRYPOINT ["/receiver-app"]
