BINARY=sentinel

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test:
	go test ./...

.PHONY: race
race:
	go test -race ./...

.PHONY: bench
bench:
	go test -bench=. ./...

.PHONY: build
build:
	go build -o bin/$(BINARY) ./cmd/sentinel

.PHONY: clean
clean:
	rm -rf bin