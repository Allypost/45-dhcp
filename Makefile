GO_BUILD_FLAGS=--ldflags '-w -s -linkmode external -extldflags "-static"' -asmflags -trimpath -race

.PHONY: all
all: client server

.PHONY: client
client:
	go build $(GO_BUILD_FLAGS) -o bin/client client.go

.PHONY: server
server:
	go build $(GO_BUILD_FLAGS) -o bin/server server.go
