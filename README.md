## Discoverx402

Minimal discovery node extension for x402 facilitators. Nodes store and gossip listings and quality signals; any client can read via a simple HTTP API.

### Prerequisites
- Go 1.22+
- Linux/macOS (SQLite via `modernc.org/sqlite`)

### Build
```bash
make build
```
The binary is written to `bin/discoverx402`.

### Run (serve API)
```bash
make run
```
By default the server listens on `:8080` and stores data in `./data/x402.db`.

### Environment variables
- `LIST_BIND` (default `:8080`): HTTP listen address
- `DB_PATH` (default `./data/x402.db`): SQLite database path
- `SEED_ONCE` (set to `1` to enable): Perform a one-time seed on startup
- `SEED_URL` (no default): Seed source URL. If not set and `SEED_ONCE=1`, seeding will return an error and be logged.

Example (run and seed once from Coinbase discovery):
```bash
SEED_ONCE=1 \
SEED_URL="https://api.cdp.coinbase.com/platform/v2/x402/discovery/resources" \
make run
```

### API

#### GET `/list`
Returns all stored listings as a JSON array (empty array `[]` when there are none).

Example:
```bash
curl -s http://localhost:8080/list | jq . | head
```

#### POST `/listings`
Upserts a single listing. Body must be a `types.Listing` JSON payload.

Example payload:
```json
{
  "resource": "https://example.com/endpoint",
  "type": "http",
  "x402Version": 1,
  "lastUpdated": "2025-01-01T00:00:00Z",
  "metadata": {"note": "example"},
  "accepts": [
    {
      "asset": "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
      "network": "solana-devnet",
      "scheme": "exact",
      "payTo": "0x0000000000000000000000000000000000000001",
      "maxAmountRequired": "1000",
      "maxTimeoutSeconds": 60,
      "mimeType": "application/json",
      "resource": "https://example.com/endpoint",
      "description": "Example accept"
    }
  ]
}
```

Insert via curl:
```bash
curl -s -X POST http://localhost:8080/listings \
  -H 'Content-Type: application/json' \
  -d @payload.json
```
Expected response:
```json
{"status":"ok"}
```
Then verify:
```bash
curl -s http://localhost:8080/list | jq .
```

### Seeding from a remote index
The seed job accepts both a raw array `[...]` and the wrapped format `{ "items": [...] }` used by Coinbase discovery.

Run once on startup:
```bash
SEED_ONCE=1 \
SEED_URL="https://api.cdp.coinbase.com/platform/v2/x402/discovery/resources" \
make run
```
Logs will report how many listings were appended, and any per-item upsert errors.

### Makefile targets
- `make build` – build the binary to `bin/`
- `make run` – run the server from sources
- `make clean` – remove `bin/`
- `make test` – simple smoke test against `/list`

### Troubleshooting
- Disk I/O errors on SQLite: we open in WAL mode by default and fall back to DELETE journal mode if needed. If you see a startup error, remove temporary SQLite files and retry:
```bash
rm -f ./data/x402.db-wal ./data/x402.db-shm
```
- Seeding shows 0 items: confirm the URL is reachable (HTTP 200) and returns either `{ "items": [...] }` or a JSON array.
- `/list` returns `[]`: no data yet; POST to `/listings` or enable seeding.