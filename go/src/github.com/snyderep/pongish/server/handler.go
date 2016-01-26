package server

import (
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/rs/xid"
	"log"
	"net/http"
)

var store = sessions.NewCookieStore([]byte("BqEKmLBysSblvwtoB4G8VjIu"))

// PongishHandlerProvider provides http handlers.
type PongishHandlerProvider struct {
	renderer       TemplateRenderer
	wsGameEndpoint string
	wsUpgrader     websocket.Upgrader
}

// NewPongishHandlerProvider creates a new PongishHandlerProvider
func NewPongishHandlerProvider(renderer TemplateRenderer, wsGameEndpoint string, wsCheckOrigin bool) *PongishHandlerProvider {
	var upgrader websocket.Upgrader
	if !wsCheckOrigin {
		upgrader = websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
	}
	return &PongishHandlerProvider{
		renderer:       renderer,
		wsGameEndpoint: wsGameEndpoint,
		wsUpgrader:     upgrader,
	}
}

func (p *PongishHandlerProvider) screenHandler(w http.ResponseWriter, r *http.Request) {
	// Get a session. Get() always returns a session, even if empty.
	session, err := store.Get(r, "pongish-a")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if session.IsNew {
		guid := xid.New()
		session.Values["id"] = guid.String()
	}

	data := make(map[string]interface{})
	data["WsGameEndpoint"] = p.wsGameEndpoint

	if err := p.renderer.renderTemplate(w, "_screen.tmpl", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (p *PongishHandlerProvider) gameHandler(w http.ResponseWriter, r *http.Request) {
	c, err := p.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("server: websocket upgrade: %s\n", err)
		return
	}

	if err := addPlayer(c); err != nil {
		log.Printf("error adding player: %s\n", err)
		return
	}
}

func (p *PongishHandlerProvider) homeHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/screen", http.StatusFound)
}
