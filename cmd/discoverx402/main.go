package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/list", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]any{
			map[string]any{
				"resource":   "https://example.com/x402/echo",
				"x402Version": 1,
				"accepts": []any{
					map[string]any{"asset":"USDC","network":"solana","scheme":"exact","payTo":"adlonymous", "maxAmountRequired":"100"},
				},
				"metadata": map[string]any{"title":"Echo"},
			},
		})
	})
	log.Println("x402node listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
