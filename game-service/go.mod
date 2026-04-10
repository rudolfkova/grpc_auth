module game

go 1.25.4

require (
	github.com/BurntSushi/toml v1.6.0
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/gorilla/websocket v1.5.3
	github.com/mlange-42/ark v0.8.0
	github.com/mlange-42/ark-serde v0.3.2
	github.com/phsym/console-slog v0.3.1
	github.com/rudolfkova/grpc_auth/pkg/gamekit v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.79.1
	world v0.0.0-00010101000000-000000000000
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/goccy/go-json v0.10.6 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace world => ../world-service

replace github.com/rudolfkova/grpc_auth/pkg/gamekit => ../pkg/gamekit
