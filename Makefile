BINARY_NAME=microblocknet
GATEWAY_NAME=gateway
CLIENT_NAME=client
NODE_SERVICE_NAME=node

build-node:
	@cd ./$(NODE_SERVICE_NAME); go build  -o ./bin/$(BINARY_NAME) -v

gate-build:
	@cd ./gateway; go build  -o ./bin/$(GATEWAY_NAME) -v

gate: gate-build
	@./gateway/bin/$(GATEWAY_NAME)

node: build-node
	@DEBUG=true ./$(NODE_SERVICE_NAME)/bin/$(BINARY_NAME)

client-build:
	@cd ./client; go build  -o ./bin/$(CLIENT_NAME) -v

client: client-build
	@./client/bin/$(CLIENT_NAME)
	
test:
	@cd ./$(NODE_SERVICE_NAME); go test -v ./... -count=1

up-all:
	@docker compose up

up:
	@docker compose up gateway node1 node2 node3 node4

down:
	@docker compose down

up-d:
	@docker compose up -d

up-b:
	@docker compose up --build

consul:
	@docker compose up -d consul-server
	@docker compose up -d consul-client

kafka:
	@docker compose up -d kafka
	
proto:
	@protoc --go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	./common/proto/*.proto
	
build: proto
	@docker build -t ghcr.io/yuriykis/microblocknet ./node
	@docker build -t ghcr.io/yuriykis/microblocknet-gateway ./gateway

clear:
	@docker images -f "dangling=true" -q | xargs -r docker rmi

.PHONY: build run test proto build-node gateway up down up-d	up-b clear