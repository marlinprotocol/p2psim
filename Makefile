all: p2psim

p2psim:
	go build -o ./build/p2psim ./cmd

test:
	go test -v ./...

cover:
	go test -v -coverpkg=./... -coverprofile=profile.cov ./...
	go tool cover -html=profile.cov
