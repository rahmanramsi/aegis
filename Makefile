.PHONY: build run-gateway run-daemon clean test vet

build:
	go build -o aegisd ./cmd/aegisd
	go build -o aegis-agent ./cmd/aegis-agent

run-gateway: build
	./aegisd

run-daemon: build
	./aegis-agent

dev:
	go run ./cmd/aegisd

clean:
	rm -f aegisd aegis-agent

test:
	go test ./... -v -count=1

test-race:
	go test ./... -v -race -count=1

vet:
	go vet ./...
