FROM golang:1.25.4-alpine AS builder
WORKDIR /src

COPY go.work ./
COPY auth-service/go.mod auth-service/go.sum ./auth-service/
COPY chat-service/go.mod chat-service/go.sum ./chat-service/
COPY gateway/go.mod gateway/go.sum ./gateway/
COPY game-service/go.mod ./game-service/
COPY world-service/go.mod world-service/go.sum ./world-service/
COPY character-service/go.mod character-service/go.sum ./character-service/
COPY pkg/gamekit/go.mod ./pkg/gamekit/
RUN go work sync && cd chat-service && go mod download

COPY auth-service ./auth-service
COPY chat-service ./chat-service
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/chat-service ./chat-service/cmd/chat-service

FROM alpine:3.20
RUN apk add --no-cache netcat-openbsd \
	&& adduser -D -u 10001 appuser
WORKDIR /app

COPY --from=builder /out/chat-service /app/chat-service

USER appuser
EXPOSE 50052
ENTRYPOINT ["/app/chat-service"]
