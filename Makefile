# Сборка сервисов.
.PHONY: build build-auth build-chat build-gateway build-game build-world build-character start start-auth start-chat start-gateway start-game docker-up docker-down docker-logs docker-reset docker-fix-iptables world-grpc-list world-grpc-ping gen-character tidy-character
ifeq ($(OS),Windows_NT)
	BIN_EXT := .exe
else
	BIN_EXT :=
endif
BIN_DIR := .
DOCKER ?= sudo docker

build-auth:
	go build -v -o $(BIN_DIR)/auth-service$(BIN_EXT) ./auth-service/cmd/auth-service
build-chat:
	go build -v -o $(BIN_DIR)/chat-service$(BIN_EXT) ./chat-service/cmd/chat-service
build-gateway:
	go build -v -o $(BIN_DIR)/gateway$(BIN_EXT) ./gateway/cmd/gateway
build-game:
	go build -v -o $(BIN_DIR)/game-service$(BIN_EXT) ./game-service/cmd/game-service
build-world:
	go build -v -o $(BIN_DIR)/world-service$(BIN_EXT) ./world-service/cmd/world-service
build-character:
	go build -v -o $(BIN_DIR)/character-service$(BIN_EXT) ./character-service/cmd/character-service

# Проверка world-service по gRPC без локального grpcurl (образ fullstorydev/grpcurl).
# Нужен запущенный сервис на localhost:50054. На Linux используется --network host.
WORLD_GRPCURL := $(DOCKER) run --rm --network host -v "$(CURDIR)/world-service/proto:/protos:ro" fullstorydev/grpcurl:latest -import-path /protos -proto world/v1/world.proto -plaintext

world-grpc-list:
	$(WORLD_GRPCURL) localhost:50054 list

world-grpc-ping:
	$(WORLD_GRPCURL) -d '{"limit":20,"offset":0}' localhost:50054 world.v1.WorldService/ListWorlds

# Запуск сервисов.
.PHONY: start
start-auth:
	$(BIN_DIR)/auth-service$(BIN_EXT) -config-path=config.toml
start-chat:
	$(BIN_DIR)/chat-service$(BIN_EXT) -config-path=config-chat.toml
start-gateway:
	$(BIN_DIR)/gateway$(BIN_EXT) -config-path=config-gateway.toml
start-game:
	$(BIN_DIR)/game-service$(BIN_EXT) -config-path=config-game.toml

# Docker-compose.
docker-up:
	$(MAKE) docker-fix-iptables
	$(DOCKER) compose up --build -d

docker-down:
	$(DOCKER) compose down --remove-orphans

docker-logs:
	$(DOCKER) compose logs -f

docker-reset:
	$(DOCKER) compose down --remove-orphans -v
	$(DOCKER) rm -f grpc-auth-redis grpc-auth-postgres grpc-auth-service grpc-chat-service grpc-gateway grpc_auth-migrate-1 grpc_auth-migrate-chat-1 grpc_auth-postgres-init-dbs-1 2>/dev/null || true

docker-fix-iptables:
	sudo iptables -t filter -N DOCKER-ISOLATION-STAGE-1 2>/dev/null || true
	sudo iptables -t filter -N DOCKER-ISOLATION-STAGE-2 2>/dev/null || true
	sudo iptables -t filter -C DOCKER-ISOLATION-STAGE-1 -j DOCKER-ISOLATION-STAGE-2 2>/dev/null || sudo iptables -t filter -A DOCKER-ISOLATION-STAGE-1 -j DOCKER-ISOLATION-STAGE-2
	sudo iptables -t filter -C DOCKER-ISOLATION-STAGE-2 -j RETURN 2>/dev/null || sudo iptables -t filter -A DOCKER-ISOLATION-STAGE-2 -j RETURN

# Линтер сервисов.
.PHONY: lint
lint-auth:
	golangci-lint run ./auth-service/...
lint-chat:
	golangci-lint run ./chat-service/...

# Миграции сервисов.
.PHONY: migrate
migrate-auth-up:
	migrate -path auth-service/migrations -database "$(DB_DSN)" up
migrate-chat-up:
	migrate -path chat-service/migrations -database "$(DB_DSN)" up

# Генерация gRPC.
.PHONY: gen
gen-auth:
	protoc -I auth-service \
	  --go_out=auth-service --go_opt=paths=source_relative \
	  --go-grpc_out=auth-service --go-grpc_opt=paths=source_relative \
	  auth-service/proto/auth/v1/auth.proto
gen-chat:
	protoc -I chat-service \
	  --go_out=chat-service --go_opt=paths=source_relative \
	  --go-grpc_out=chat-service --go-grpc_opt=paths=source_relative \
	  chat-service/proto/chat/v1/chat.proto
gen-world:
	protoc -I world-service \
	  --go_out=world-service --go_opt=paths=source_relative \
	  --go-grpc_out=world-service --go-grpc_opt=paths=source_relative \
	  world-service/proto/world/v1/world.proto
gen-character:
	protoc -I character-service \
	  --go_out=character-service --go_opt=paths=source_relative \
	  --go-grpc_out=character-service --go-grpc_opt=paths=source_relative \
	  character-service/proto/character/v1/character.proto

# go mod tidy по сервисам.
.PHONY: tidy
tidy-auth:
	cd auth-service && go mod tidy
tidy-chat:
	cd chat-service && go mod tidy
tidy-gateway:
	cd gateway && go mod tidy
tidy-game:
	cd game-service && go mod tidy
tidy-world:
	cd world-service && go mod tidy
tidy-character:
	cd character-service && go mod tidy

# Форматирование всего репозитория.
.PHONY: gofmt
gofmt:
	gofmt -w -s .

# Тесты сервисов.
.PHONY: test
test-auth:
	cd auth-service && go test -v -race -timeout 30s ./...
test-chat:
	cd chat-service && go test -v -race -timeout 30s ./...
# Интеграционные тесты сервисов.
test-auth-integration:
	cd auth-service && go test ./... -tags=integration -v
test-chat-integration:
	cd chat-service && go test ./... -tags=integration -v

# Моки для интерфейсов.
.PHONY: mocks
mocks-auth:
	cd auth-service && mockery --name=Auth --dir=./internal/grpc/auth --output=./mocks/auth --outpkg=mocks
	cd auth-service && mockery --name=UserRepository --dir=./internal/repository --output=./mocks/repository --outpkg=mocks
	cd auth-service && mockery --name=SessionRepository --dir=./internal/repository --output=./mocks/repository --outpkg=mocks
	cd auth-service && mockery --name=Cache --dir=./internal/repository --output=./mocks/repository --outpkg=mocks
	cd auth-service && mockery --name=TokenProvider --dir=./provider --output=./mocks/provider --outpkg=mocks
	cd auth-service && mockery --name=AuthServiceServer --dir=./proto/auth/v1 --output=./mocks/proto/auth/v1 --outpkg=mocks
	cd auth-service && mockery --name=AuthServiceClient --dir=./proto/auth/v1 --output=./mocks/proto/auth/v1 --outpkg=mocks
	cd auth-service && mockery --name=UnsafeAuthServiceServer --dir=./proto/auth/v1 --output=./mocks/proto/auth/v1 --outpkg=mocks