package main

import (
	"context"
	"log"
	"os"

	httpapi "github.com/adlonymous/discoverx402/internal/http"
	"github.com/adlonymous/discoverx402/internal/state"
	"github.com/adlonymous/discoverx402/internal/types"
)

func main() {
	addr := getenv("LIST_BIND", ":8080")
	dbPath := getenv("DB_PATH", "./data/x402.db")

	repo, err := state.OpenSQLite(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Close()

	// seed a single listing so /list returns something
	_ = repo.Upsert(context.Background(), sample())

	srv := httpapi.New(addr, repo)
	log.Fatal(srv.Start())
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func sample() types.Listing {
	return types.Listing{
		Accepts: []types.Accept{{
			Asset:             "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
			Extra:             &types.AcceptExtra{Name: "USD Coin", Version: "2"},
			MaxAmountRequired: "200", MaxTimeoutSeconds: 60,
			Network: "base", Scheme: "exact",
			PayTo:        "0xa2477E16dCB42E2AD80f03FE97D7F1a1646cd1c0",
			Resource:     "https://api.example.com/x402/last_sold",
			OutputSchema: &types.OutputSchema{Input: types.OutputInput{Method: "GET", Type: "http"}, Output: nil},
		}},
		LastUpdated: "2025-08-09T01:07:04.005Z",
		Metadata:    map[string]any{},
		Resource:    "https://api.prixe.io/x402/last_sold",
		Type:        "http",
		X402Version: 1,
	}
}
