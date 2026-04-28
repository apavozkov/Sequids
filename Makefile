.PHONY: test build proto

test:
	go test ./...

build:
	mkdir -p bin
	go build -o ./bin/central ./cmd/central
	go build -o ./bin/worker ./cmd/worker
	go build -o ./bin/sequidsctl ./cmd/sequidsctl

proto:
	protoc --go_out=. --go-grpc_out=. api/proto/orchestrator.proto
