FROM golang:1.25.4-alpine AS builder
WORKDIR /src

COPY gateway/go.mod gateway/go.sum ./gateway/
COPY character-service/go.mod character-service/go.sum ./character-service/

WORKDIR /src/gateway
RUN go mod download

COPY gateway/ .
COPY character-service/ ../character-service/

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/gateway ./cmd/gateway

FROM alpine:3.20
RUN adduser -D -u 10001 appuser
WORKDIR /app

COPY --from=builder /out/gateway /app/gateway

USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/gateway"]
