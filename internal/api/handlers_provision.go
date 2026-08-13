package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// handleProvisionPerforce proxies a creator's Perforce access request to
// hydraperforceprovision, forwarding their iamnim session. The portal holds no
// credentials: hydraperforceprovision validates the session against iamnim and
// mints the access. This is a thin authenticated proxy — the real auth,
// membership check and provisioning all happen downstream.
func (s *Server) handleProvisionPerforce(w http.ResponseWriter, r *http.Request) {
	if s.provisionPerforceURL == "" {
		writeProvisionError(w, http.StatusServiceUnavailable, "perforce provisioning is not configured")
		return
	}

	session := iamnimSession(r)
	if session == "" {
		writeProvisionError(w, http.StatusUnauthorized, "sign in with iamnim first")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16))
	if err != nil {
		writeProvisionError(w, http.StatusBadRequest, "unreadable request body")
		return
	}

	up, err := http.NewRequestWithContext(r.Context(), http.MethodPost,
		s.provisionPerforceURL+"/api/v1/provision", bytes.NewReader(body))
	if err != nil {
		writeProvisionError(w, http.StatusInternalServerError, "could not build upstream request")
		return
	}
	up.Header.Set("Content-Type", "application/json")
	up.Header.Set("X-Iamnim-Session", session)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(up)
	if err != nil {
		writeProvisionError(w, http.StatusBadGateway, "provisioning service unreachable")
		return
	}
	defer resp.Body.Close()

	// Pass the upstream status and body straight back to the caller.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, io.LimitReader(resp.Body, 1<<20))
}

// iamnimSession pulls the creator's session from the iamnim_session cookie, the
// X-Iamnim-Session header, or a ?token= param.
func iamnimSession(r *http.Request) string {
	if c, err := r.Cookie("iamnim_session"); err == nil && c.Value != "" {
		return c.Value
	}
	if h := r.Header.Get("X-Iamnim-Session"); h != "" {
		return h
	}
	return r.URL.Query().Get("token")
}

func writeProvisionError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
