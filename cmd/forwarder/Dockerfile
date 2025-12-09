# Stage 1: Build
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build only the forwarder component
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /forwarder-app ./cmd/forwarder

# Stage 2: Final Image
FROM alpine:3.18
COPY --from=builder /forwarder-app /forwarder-app
ENTRYPOINT ["/forwarder-app"]
