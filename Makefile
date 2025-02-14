# Client and Server source directories
CLIENT_DIR=client
SERVER_DIR=server

# Output binary names
CLIENT_OUT_LINUX=veiltransfer_client_linux
CLIENT_OUT_WINDOWS=veiltransfer_client_windows.exe
CLIENT_OUT_MACOS=veiltransfer_client_macos

SERVER_OUT_LINUX=veiltransfer_server_linux
SERVER_OUT_WINDOWS=veiltransfer_server_windows.exe
SERVER_OUT_MACOS=veiltransfer_server_macos

# Default target
all: client server

# Build Client
client: client-linux client-windows client-macos

client-linux:
	cd $(CLIENT_DIR) && go mod tidy && go mod download && GOOS=linux GOARCH=amd64 go build -o ../$(CLIENT_OUT_LINUX) ./main.go
	strip $(CLIENT_OUT_LINUX)

client-windows:
	cd $(CLIENT_DIR) && go mod tidy && go mod download && GOOS=windows GOARCH=amd64 go build -o ../$(CLIENT_OUT_WINDOWS) ./main.go

client-macos:
	cd $(CLIENT_DIR) && go mod tidy && go mod download && GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o ../$(CLIENT_OUT_MACOS) ./main.go

# Build Server
server: server-linux server-windows server-macos

server-linux:
	cd $(SERVER_DIR) && go mod tidy && go mod download && GOOS=linux GOARCH=amd64 go build -o ../$(SERVER_OUT_LINUX) ./main.go
	strip $(SERVER_OUT_LINUX)

server-windows:
	cd $(SERVER_DIR) && go mod tidy && go mod download && GOOS=windows GOARCH=amd64 go build -o ../$(SERVER_OUT_WINDOWS) ./main.go

server-macos:
	cd $(SERVER_DIR) && go mod tidy && go mod download && GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o ../$(SERVER_OUT_MACOS) ./main.go

# Clean the build artifacts
clean:
	rm -f $(CLIENT_OUT_LINUX) $(CLIENT_OUT_WINDOWS) $(CLIENT_OUT_MACOS)
	rm -f $(SERVER_OUT_LINUX) $(SERVER_OUT_WINDOWS) $(SERVER_OUT_MACOS)

# List files
list:
	ls -la

.PHONY: all client server clean list
