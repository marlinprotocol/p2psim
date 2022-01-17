all: p2psim

p2psim:
	go build -o ./build/p2psim ./cmd

test:
	go test -v ./...

cover:
	go test -v ./... -coverprofile cover.out
	go tool cover -html=cover.out
