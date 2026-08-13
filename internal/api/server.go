package api

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/cederikdotcom/hydramancer/internal/web"
)

type Server struct {
	templates           *template.Template
	mux                 *http.ServeMux
	provisionPerforceURL string
}

// NewServer builds the portal. provisionPerforceURL is the hydraperforceprovision
// base URL the portal proxies to; empty disables the /provision/perforce route.
func NewServer(provisionPerforceURL string) *Server {
	srv := &Server{provisionPerforceURL: provisionPerforceURL}

	srv.templates = template.Must(
		template.New("").ParseFS(web.TemplateFS, "templates/*.html"),
	)

	srv.mux = http.NewServeMux()
	srv.registerRoutes()

	return srv
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /{$}", s.webLanding)
	s.mux.HandleFunc("GET /deploy", s.webDeploy)
	s.mux.HandleFunc("GET /experience", s.webExperience)
	s.mux.HandleFunc("GET /api/v1/health", s.apiHealth)
	s.mux.HandleFunc("POST /api/v1/provision/perforce", s.handleProvisionPerforce)
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

// webDeploy serves the developer quickstart: how to ship a service onto Hydra.
func (s *Server) webDeploy(w http.ResponseWriter, r *http.Request) {
	s.templates.ExecuteTemplate(w, "deploy.html", nil)
}

// webExperience serves the creator quickstart: how to publish an Unreal
// experience through the draft -> staging -> live lifecycle.
func (s *Server) webExperience(w http.ResponseWriter, r *http.Request) {
	s.templates.ExecuteTemplate(w, "experience.html", nil)
}
