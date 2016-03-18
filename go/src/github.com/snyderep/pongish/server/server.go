package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

// HandlerProvider provides all HTTP handlers.
type HandlerProvider interface {
	homeHandler(w http.ResponseWriter, r *http.Request)
	screenHandler(w http.ResponseWriter, r *http.Request)
	gameHandler(w http.ResponseWriter, r *http.Request)
}

// TemplateRenderer renders templates (of course).
type TemplateRenderer interface {
	renderTemplate(w http.ResponseWriter, templateName string, data map[string]interface{}) error
}

// Server wraps an http server.
type Server struct {
	Address      string
	Provider     HandlerProvider
	StaticPrefix string
	StaticRoot   string
}

// Listen starts listening for HTTP traffic.
func (s *Server) Listen() error {
	router, err := s.setupRoutes()
	if err != nil {
		return err
	}

	// handle static content
	http.Handle(s.StaticPrefix, http.StripPrefix(s.StaticPrefix, http.FileServer(http.Dir(s.StaticRoot))))

	// handle everything else
	http.Handle("/", router)

	return http.ListenAndServe(s.Address, nil)
}

func (s *Server) setupRoutes() (*mux.Router, error) {
	r := mux.NewRouter()

	r.HandleFunc("/", s.Provider.homeHandler)

	// game screen template handler
	r.HandleFunc("/screen", s.Provider.screenHandler)

	// game websocket handler
	r.HandleFunc("/game", s.Provider.gameHandler)

	return r, nil
}
