BINARY_NAME=microblocknet

build-binary:
	@go build -o bin/$(BINARY_NAME) -v

run: build-binary
	@DEBUG=true ./bin/$(BINARY_NAME)

test:
	@go test -v ./...

up:
	@docker compose up

down:
	@docker compose down

up-d:
	@docker compose up -d

proto:
	@protoc --go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	proto/*.proto
	
build: proto
	@docker build -t microblocknet .

.PHONY: build run test proto