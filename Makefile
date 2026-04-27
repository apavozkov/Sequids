.PHONY: test build proto

test:
	go test ./...

build:
	go build ./cmd/central ./cmd/worker ./cmd/sequidsctl

proto:
	protoc --go_out=. --go-grpc_out=. api/proto/orchestrator.proto
