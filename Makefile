MAIN_FILE := ./cmd/balancer/main.go

.PHONY: run
run:
	go run ${MAIN_FILE}

.PHONY: lint
lint:
	golangci-lint run --show-stats