package p2p

import (
	"encoding/json"
	"net/http"
)

type apiFunc func(http.ResponseWriter, *http.Request) error

func makeHttpHandleFunc(fn apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			writeJson(w, http.StatusBadRequest, map[string]any{"err": err.Error()})
		}
	}
}

func writeJson(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

type ApiServer struct {
	listenAddr string
	game       *Game
}

func NewApiServer(addr string, game *Game) *ApiServer {
	return &ApiServer{
		listenAddr: addr,
		game:       game,
	}
}

func (s *ApiServer) Run() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /ready", makeHttpHandleFunc(s.handlePlayersReady))
	mux.HandleFunc("GET /fold", makeHttpHandleFunc(s.handlePlayersFold))
	http.ListenAndServe(s.listenAddr, mux)
}

func (s *ApiServer) handlePlayersReady(w http.ResponseWriter, r *http.Request) error {
	s.game.SetReady()
	return writeJson(w, http.StatusOK, map[string]any{"Game server": s.game.currentStatus})
}

func (s *ApiServer) handlePlayersFold(w http.ResponseWriter, r *http.Request) error {
	s.game.Fold()
	return writeJson(w, http.StatusOK, map[string]any{"Game server": s.game.currentStatus})
}
