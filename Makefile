BINARY_NAME=microblocknet

build:
	@go build -o bin/$(BINARY_NAME) -v

run: build
	@./bin/$(BINARY_NAME)

test:
	@go test -v ./...

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/types.proto

.PHONY: build run test proto