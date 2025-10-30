package mirror

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/adlonymous/discoverx402/internal/state"
	"github.com/adlonymous/discoverx402/internal/types"
)

func SeedOnce(ctx context.Context, url string, repo *state.Repo) error {
	if url == "" {
		return fmt.Errorf("seed: empty SEED_URL")
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("seed: non-200 status %d from %s: %s", resp.StatusCode, url, string(body))
	}

	body, _ := io.ReadAll(resp.Body)

	var items []types.Listing

	var wrapped struct {
		Items []types.Listing `json:"items"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil && len(wrapped.Items) > 0 {
		items = wrapped.Items
	} else {
		if err := json.Unmarshal(body, &items); err != nil {
			return fmt.Errorf("seed: failed to decode response from %s: %w", url, err)
		}
	}

	n := 0
	for _, it := range items {
		if err := repo.Upsert(ctx, it); err == nil {
			n++
		} else {
			log.Printf("seed: upsert failed for resource %s: %v", it.Resource, err)
		}
	}
	log.Printf("seed: appended %d listings from %s", n, url)
	return nil
}
