APP := discoverx402
BIN := bin/$(APP)

.DEFAULT_GOAL := build

build:
	go build -o $(BIN) ./cmd/$(APP)

run: 
	go run ./cmd/$(APP)

clean:
	rm -rf bin

clean-db:
	rm -rf data/x402.db

test:
	curl -X GET http://localhost:8080/list
