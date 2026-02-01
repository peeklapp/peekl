gen-api-doc:
	swag init -g pkg/api/api.go

build-all: build-server-linux-amd64 build-agent-linux-amd64 build-server-linux-arm64 build-agent-linux-arm64

OUTPUT_DIR = ./output

build-server-linux-amd64:
	env GOOS=linux GOARCH=amd64 go build -o $(OUTPUT_DIR)/peekl-server-linux-amd64 cmd/server/main.go

build-agent-linux-amd64:
	env GOOS=linux GOARCH=amd64 go build -o $(OUTPUT_DIR)/peekl-agent-linux-amd64 cmd/agent/main.go

# ARM64
build-server-linux-arm64:
	env GOOS=linux GOARCH=arm64 go build -o $(OUTPUT_DIR)/peekl-server-linux-arm64 cmd/server/main.go

build-agent-linux-arm64:
	env GOOS=linux GOARCH=arm64 go build -o $(OUTPUT_DIR)/peekl-agent-linux-arm64 cmd/agent/main.go
