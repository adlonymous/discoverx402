package httpserver

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/adlonymous/discoverx402/internal/state"
	"github.com/adlonymous/discoverx402/internal/types"
)

type Server struct {
	addr string
	mux  *http.ServeMux
	repo *state.Repo
}

func New(addr string, repo *state.Repo) *Server {
	s := &Server{addr: addr, mux: http.NewServeMux(), repo: repo}
	s.mux.HandleFunc("/list", s.handleList)
	s.mux.HandleFunc("/listings", s.handleUpsertListing)
	s.mux.HandleFunc("/healthz", s.handleHealth)
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
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }
