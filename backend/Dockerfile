FROM golang:1.24-alpine AS builder

WORKDIR /src/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux go build -o /migrate ./cmd/migrate

FROM alpine:3.20

RUN apk add --no-cache ca-certificates netcat-openbsd

WORKDIR /app

COPY --from=builder /server .
COPY --from=builder /migrate .
COPY backend/docker-entrypoint.sh .
COPY sql /sql

ENV SQL_DIR=/sql

RUN chmod +x docker-entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["./docker-entrypoint.sh"]
