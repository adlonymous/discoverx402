package httpserver

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	p2p "github.com/adlonymous/discoverx402/internal/p2p"
	"github.com/adlonymous/discoverx402/internal/state"
	"github.com/adlonymous/discoverx402/internal/types"
)

type Server struct {
	addr string
	mux  *http.ServeMux
	repo *state.Repo
	node *p2p.Node
}

func New(addr string, repo *state.Repo, node *p2p.Node) *Server {
	s := &Server{addr: addr, mux: http.NewServeMux(), repo: repo, node: node}
	s.mux.HandleFunc("/list", s.handleList)
	s.mux.HandleFunc("/listings", s.handleUpsertListing)
	s.mux.HandleFunc("/healthz", s.handleHealth)
	s.mux.HandleFunc("/p2p/rebroadcast", s.handleRebroadcast)
	s.mux.HandleFunc("/p2p/peers", s.handleListPeers)
	return s
}

func (s *Server) Start() error {
	log.Println("x402node listening on", s.addr)
	return http.ListenAndServe(s.addr, s.mux)
}

func (s *Server) handleList(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	listings, err := s.repo.List(context.Background())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	_ = json.NewEncoder(w).Encode(listings)
}

func (s *Server) handleRebroadcast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if s.node == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	l, err := s.repo.List(r.Context())
	if err != nil {
		for _, item := range l {
			_ = s.node.Publish(r.Context(), item)
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleUpsertListing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var l types.Listing
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&l); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.repo.Upsert(context.Background(), l); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if s.node != nil {
		if err := s.node.Publish(context.Background(), l); err != nil {
			log.Printf("failed to publish listing to gossip: %v", err)
		}
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }

func (s *Server) handleListPeers(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.node == nil {
		_ = json.NewEncoder(w).Encode(map[string]any{"peers": []string{}, "count": 0})
		return
	}
	peers := s.node.ListPeers()
	peerStrs := make([]string, len(peers))
	for i, p := range peers {
		peerStrs[i] = p.String()
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"peers": peerStrs, "count": len(peers)})
}
