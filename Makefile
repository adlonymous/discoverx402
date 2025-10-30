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

test-populate:
	@printf '%b' '{\n  "resource": "https://example.com/echo",\n  "type": "http",\n  "x402Version": 1,\n  "lastUpdated": "2025-01-01T00:00:00Z",\n  "metadata": {},\n  "accepts": [\n    {\n      "asset": "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",\n      "network": "base",\n      "scheme": "exact",\n      "payTo": "0x0000000000000000000000000000000000000001",\n      "maxAmountRequired": "1000",\n      "maxTimeoutSeconds": 60,\n      "mimeType": "application/json",\n      "resource": "https://example.com/echo",\n      "description": "Sample resource"\n    }\n  ]\n}' | curl -s -X POST http://localhost:8080/listings -H 'Content-Type: application/json' --data-binary @-
