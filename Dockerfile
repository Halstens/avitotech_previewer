# Dockerfile
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app


COPY go.mod go.sum ./
RUN go mod download


COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates && \
    addgroup -S app && adduser -S app -G app

WORKDIR /home/app

COPY --from=builder --chown=app:app /app/main .
COPY --from=builder --chown=app:app /app/migrations ./migrations

USER app

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./main"]