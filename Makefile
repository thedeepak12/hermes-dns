.PHONY: build run

build:
	go build -o bin/hermes ./cmd/hermes/main.go

run:
	go run ./cmd/hermes/main.go
