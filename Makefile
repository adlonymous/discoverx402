APP := discoverx402
BIN := bin/$(APP)

.DEFAULT_GOAL := build

build:
	go build -o $(BIN) ./cmd/$(APP)

run: 
	go run ./cmd/$(APP)

clean:
	rm -rf bin