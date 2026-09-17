# Stage 1: Build the application
FROM golang:1.24-alpine AS builder

RUN apk update && apk add --no-cache \
    ca-certificates \
    git \
    && rm -rf /var/cache/apk/*

RUN adduser -D -s /bin/sh appuser

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main .

# Stage 2: Create a minimal production image
FROM alpine:3.19

RUN apk update && apk add --no-cache \
    ca-certificates \
    tzdata \
    && rm -rf /var/cache/apk/*

RUN adduser -D -s /bin/sh -u 1001 appuser

WORKDIR /app

COPY --from=builder /app/main .

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 3001

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3001/api/v1/health || exit 1

ENV GIN_MODE=release
ENV CGO_ENABLED=0
ENV BACKEND_PORT=3001
ENV PREFIX=/api/v1

CMD ["./main"]
