FROM golang:1.25.4-alpine AS builder
WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/game-service ./cmd/game-service

FROM alpine:3.20
RUN adduser -D -u 10001 appuser
WORKDIR /app

COPY --from=builder /out/game-service /app/game-service

USER appuser
EXPOSE 50053
ENTRYPOINT ["/app/game-service"]
