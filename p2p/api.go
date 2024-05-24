package p2p

import (
	"encoding/json"
	"net/http"
	"strconv"
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
	mux        *http.ServeMux
}

func NewApiServer(addr string, game *Game) *ApiServer {
	return &ApiServer{
		listenAddr: addr,
		game:       game,
		mux:        http.NewServeMux(),
	}
}

func (s *ApiServer) Run() {
	s.mux.HandleFunc("GET /ready", makeHttpHandleFunc(s.handlePlayersReady))
	s.mux.HandleFunc("GET /fold", makeHttpHandleFunc(s.handlePlayersFold))
	s.mux.HandleFunc("GET /check", makeHttpHandleFunc(s.handlePlayersChecked))
	s.mux.HandleFunc("GET /bet/{value}", makeHttpHandleFunc(s.handlePlayersBet))

	http.ListenAndServe(s.listenAddr, s.mux)
}

func (s *ApiServer) handlePlayersReady(w http.ResponseWriter, r *http.Request) error {
	s.game.SetReady()
	return writeJson(w, http.StatusOK, "READY")
}

func (s *ApiServer) handlePlayersBet(w http.ResponseWriter, r *http.Request) error {
	val := r.PathValue("value")
	intVal, _ := strconv.Atoi(val)
	if err := s.game.TakeAction(PlayerActionBet, intVal); err != nil {
		return err
	}
	return writeJson(w, http.StatusOK, "READY")
}

func (s *ApiServer) handlePlayersChecked(w http.ResponseWriter, r *http.Request) error {
	if err := s.game.TakeAction(PLayerActionCheck, 0); err != nil {
		return err
	}
	return writeJson(w, http.StatusOK, "CHECKED")
}

func (s *ApiServer) handlePlayersFold(w http.ResponseWriter, r *http.Request) error {
	if err := s.game.TakeAction(PlayerActionFold, 0); err != nil {
		return err
	}
	return writeJson(w, http.StatusOK, "FOLDED")
}
