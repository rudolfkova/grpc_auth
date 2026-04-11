FROM golang:1.25.4-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/world-service ./cmd/world-service

FROM alpine:3.20
RUN apk add --no-cache netcat-openbsd \
	&& adduser -D -u 10001 appuser
WORKDIR /app

COPY --from=builder /out/world-service /app/world-service

USER appuser
EXPOSE 50054
ENTRYPOINT ["/app/world-service"]
