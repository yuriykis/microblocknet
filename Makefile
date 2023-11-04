BINARY_NAME=microblocknet
GATEWAY_NAME=gateway
CLIENT_NAME=client
NODE_SERVICE_NAME=node

build-binary:
	@cd ./$(NODE_SERVICE_NAME); go build  -o ./bin/$(BINARY_NAME) -v

run: build-binary
	@DEBUG=true ./$(NODE_SERVICE_NAME)/bin/$(BINARY_NAME)

gate-build:
	@cd ./gateway; go build  -o ./bin/$(GATEWAY_NAME) -v

gate: gate-build
	@./gateway/bin/$(GATEWAY_NAME)

client-build:
	@cd ./client; go build  -o ./bin/$(CLIENT_NAME) -v

client: client-build
	@./client/bin/$(CLIENT_NAME)
	
test:
	@cd ./$(NODE_SERVICE_NAME); go test -v ./... -count=1

up:
	@docker compose up

down:
	@docker compose down

up-d:
	@docker compose up -d

up-b:
	@docker compose up --build
	
proto:
	@protoc --go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	./common/proto/*.proto
	
build: proto
	@docker build -t microblocknet ./node
	@docker build -t microblocknet-gateway ./gateway

clear:
	@docker rmi $(docker images -f "dangling=true" -q)

.PHONY: build run test proto build-binary gateway up down up-d	up-b clear