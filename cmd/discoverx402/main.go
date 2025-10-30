package main

import (
	"context"
	"log"
	"os"

	httpapi "github.com/adlonymous/discoverx402/internal/http"
	"github.com/adlonymous/discoverx402/internal/mirror"
	"github.com/adlonymous/discoverx402/internal/state"
)

func main() {
	addr := getenv("LIST_BIND", ":8080")
	dbPath := getenv("DB_PATH", "./data/x402.db")

	repo, err := state.OpenSQLite(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Close()

	seedURL := getenv("SEED_URL", "")
	if getenv("SEED_ONCE", "") == "1" {
		if err := mirror.SeedOnce(context.Background(), seedURL, repo); err != nil {
			log.Printf("seed failed: %v", err)
		}
	}

	srv := httpapi.New(addr, repo)
	log.Fatal(srv.Start())
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
