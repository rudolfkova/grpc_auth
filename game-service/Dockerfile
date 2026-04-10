FROM golang:1.25.4-alpine AS builder
WORKDIR /src

COPY go.work ./
COPY auth-service/go.mod auth-service/go.sum ./auth-service/
COPY chat-service/go.mod chat-service/go.sum ./chat-service/
COPY gateway/go.mod gateway/go.sum ./gateway/
COPY game-service/go.mod game-service/go.sum ./game-service/
COPY world-service/go.mod world-service/go.sum ./world-service/
RUN go work sync && cd game-service && go mod download

COPY game-service ./game-service
COPY world-service ./world-service
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/game-service ./game-service/cmd/game-service

FROM alpine:3.20
RUN adduser -D -u 10001 appuser
WORKDIR /app

COPY --from=builder /out/game-service /app/game-service

USER appuser
EXPOSE 50053
ENTRYPOINT ["/app/game-service"]
