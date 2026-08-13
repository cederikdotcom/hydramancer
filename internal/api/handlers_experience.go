package api

import (
	"net/http"
	"net/url"
)

const sessionCookie = "iamnim_session"

// handleLogin sends the creator to iamnim to sign in, asking iamnim to return
// them to the portal with a session token. iamnim's /login threads redirect_uri
// through every sign-in method (Google, email), and .experiencenet.com is on its
// redirect allow-list.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	ret := "https://" + s.domain + "/experience/authed"
	dest := s.iam.BaseURL() + "/login?redirect_uri=" + url.QueryEscape(ret)
	http.Redirect(w, r, dest, http.StatusFound)
}

// handleAuthed captures the token iamnim appends on return and stores it in the
// portal's own session cookie, then sends the creator back to /experience.
func (s *Server) handleAuthed(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     sessionCookie,
			Value:    token,
			Path:     "/",
			MaxAge:   86400, // matches iamnim's 24h session
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})
	}
	http.Redirect(w, r, "/experience", http.StatusSeeOther)
}

// handleLogout clears the portal session cookie.
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/experience", http.StatusSeeOther)
}
