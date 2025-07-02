# minter/Dockerfile
FROM golang:1.23.10

WORKDIR /app

COPY . .

RUN go mod tidy && go build -o minter main.go

CMD ["./minter"]
