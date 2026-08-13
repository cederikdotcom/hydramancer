package api

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/cederikdotcom/hydramancer/internal/iamnim"
	"github.com/cederikdotcom/hydramancer/internal/web"
)

type Server struct {
	templates            *template.Template
	mux                  *http.ServeMux
	provisionPerforceURL string
	iam                  *iamnim.Client
	domain               string
}

// NewServer builds the portal.
//   - provisionPerforceURL: hydraperforceprovision base URL the portal proxies to
//     (empty disables provisioning).
//   - iamnimURL: identity service used for sign-in and identity/memberships.
//   - domain: the portal's own public domain, for building the sign-in return URL.
func NewServer(provisionPerforceURL, iamnimURL, domain string) *Server {
	srv := &Server{
		provisionPerforceURL: provisionPerforceURL,
		iam:                  iamnim.New(iamnimURL),
		domain:               domain,
	}

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
	s.mux.HandleFunc("GET /experience/login", s.handleLogin)
	s.mux.HandleFunc("GET /experience/authed", s.handleAuthed)
	s.mux.HandleFunc("GET /experience/logout", s.handleLogout)
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

// experienceData drives the sign-in / access panel on /experience.
type experienceData struct {
	SignedIn bool
	Email    string
	Orgs     []iamnim.Membership
	CanProvision bool
}

// webExperience serves the creator quickstart. If the creator has an iamnim
// session (via the portal's cookie), it also renders the "Get Perforce access"
// panel with their org memberships to choose from.
func (s *Server) webExperience(w http.ResponseWriter, r *http.Request) {
	data := experienceData{CanProvision: s.provisionPerforceURL != ""}
	if session := iamnimSession(r); session != "" {
		if user, err := s.iam.Me(session); err == nil {
			data.SignedIn = true
			data.Email = user.Email
			if orgs, err := s.iam.Memberships(session); err == nil {
				data.Orgs = orgs
			}
		}
	}
	s.templates.ExecuteTemplate(w, "experience.html", data)
}
