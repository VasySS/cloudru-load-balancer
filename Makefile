MAIN_FILE := ./cmd/balancer/main.go

.PHONY: run
run:
	go run ${MAIN_FILE}

.PHONY: lint
lint:
	golangci-lint run --show-stats

.PHONY: load
load:
	ab -n 1000 -c 100 http://localhost:8080/
