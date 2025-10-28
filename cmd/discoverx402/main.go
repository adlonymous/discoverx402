package main

import (
	"log"
	"os"

	httpapi "github.com/adlonymous/discoverx402/internal/http"
)

func main() {
	addr := getenv("LIST_BIND", ":8080")
	srv := httpapi.New(addr)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
