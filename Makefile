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

test-podman:
	podman rmi peekl-test:latest || true
	podman build -t peekl-test:latest -f tests.Dockerfile
	podman run --rm --user=root peekl-test:latest
