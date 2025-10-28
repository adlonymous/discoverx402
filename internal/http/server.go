package httpserver

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/adlonymous/discoverx402/internal/types"
)

type Server struct {
	addr     string
	mux      *http.ServeMux
	listings []types.Listing
}

func New(addr string) *Server {
	s := &Server{
		addr: addr,
		mux:  http.NewServeMux(),
		listings: []types.Listing{{
			Accepts: []types.Accept{{
				Asset:             "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
				Description:       "",
				Extra:             &types.AcceptExtra{Name: "USD Coin", Version: "2"},
				MaxAmountRequired: "200",
				MaxTimeoutSeconds: 60,
				MimeType:          "",
				Network:           "base",
				OutputSchema: &types.OutputSchema{
					Input:  types.OutputInput{Method: "GET", Type: "http"},
					Output: nil,
				},
				PayTo:    "0xa2477E16dCB42E2AD80f03FE97D7F1a1646cd1c0",
				Resource: "https://api.example.com/x402/last_sold",
				Scheme:   "exact",
			}},
			LastUpdated: "2025-08-09T01:07:04.005Z",
			Metadata:    map[string]any{},
			Resource:    "https://api.prixe.io/x402/last_sold",
			Type:        "http",
			X402Version: 1,
		}},
	}

	s.mux.HandleFunc("/list", s.handleList)
	s.mux.HandleFunc("/healthz", s.handleHealth)

	return s
}

func (s *Server) Start() error {
	log.Println("x402node listening on", s.addr)
	return http.ListenAndServe(s.addr, s.mux)
}

func (s *Server) handleList(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.listings)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
