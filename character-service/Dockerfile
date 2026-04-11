FROM golang:1.25.4-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/character-service ./cmd/character-service

FROM alpine:3.20
RUN adduser -D -u 10001 appuser
WORKDIR /app

COPY --from=builder /out/character-service /app/character-service

USER appuser
EXPOSE 50055
ENTRYPOINT ["/app/character-service"]
