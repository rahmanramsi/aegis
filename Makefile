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
	go test ./... -count=1

test-cover:
	go test ./... -cover -count=1

test-race:
	go test ./... -race -count=1

smoke:
	bash scripts/smoke-test.sh

check: vet test
	@echo "All checks passed"

vet:
	go vet ./...

build-web:
	cd web && npm run build

dev:
	go run ./cmd/aegisd
