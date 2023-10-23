BINARY_NAME=microblocknet
NODE_SERVICE_NAME=node

build-binary:
	@cd ./$(NODE_SERVICE_NAME); go build  -o ./bin/$(BINARY_NAME) -v

run: build-binary
	@DEBUG=true ./$(NODE_SERVICE_NAME)/bin/$(BINARY_NAME)

test:
	@cd ./$(NODE_SERVICE_NAME); go test -v ./... -count=1

up:
	@docker compose up

down:
	@docker compose down

up-d:
	@docker compose up -d

proto:
	@protoc --go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	./$(NODE_SERVICE_NAME)/proto/*.proto
	
build: proto
	@docker build -t microblocknet ./node

.PHONY: build run test proto