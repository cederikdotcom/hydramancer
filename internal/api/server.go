package api

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/cederikdotcom/hydramancer/internal/web"
)

type Server struct {
	templates *template.Template
	mux       *http.ServeMux
}

func NewServer() *Server {
	srv := &Server{}

	srv.templates = template.Must(
		template.New("").ParseFS(web.TemplateFS, "templates/*.html"),
	)

	srv.mux = http.NewServeMux()
	srv.registerRoutes()

	return srv
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /{$}", s.webLanding)
	s.mux.HandleFunc("GET /api/v1/health", s.apiHealth)
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) apiHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) webLanding(w http.ResponseWriter, r *http.Request) {
	s.templates.ExecuteTemplate(w, "index.html", nil)
}
