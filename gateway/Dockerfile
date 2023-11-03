FROM golang:1.21.3

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o bin/gateway .

CMD ["./bin/gateway"]