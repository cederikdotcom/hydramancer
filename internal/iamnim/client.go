// Package iamnim is a small client for the identity service, used by the portal
// to show the signed-in creator and their org memberships.
package iamnim

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// User is GET /api/me.
type User struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

// Membership is one entry of GET /api/me/memberships.
type Membership struct {
	OrganizationSlug string `json:"organization_slug"`
	OrganizationName string `json:"organization_name"`
}

// Client talks to iamnim, forwarding a creator's session.
type Client struct {
	baseURL string
	http    *http.Client
}

// New returns a client. baseURL e.g. https://iamnim.com.
func New(baseURL string) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), http: &http.Client{Timeout: 10 * time.Second}}
}

// BaseURL is the configured origin (for building sign-in links).
func (c *Client) BaseURL() string { return c.baseURL }

// Me resolves the identity behind a session. A non-nil user means valid.
func (c *Client) Me(session string) (*User, error) {
	var u User
	if err := c.get("/api/me", session, &u); err != nil {
		return nil, err
	}
	if u.UserID == "" && u.Email == "" {
		return nil, fmt.Errorf("iamnim: empty identity")
	}
	return &u, nil
}

// Memberships lists the session user's orgs.
func (c *Client) Memberships(session string) ([]Membership, error) {
	var m []Membership
	if err := c.get("/api/me/memberships", session, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *Client) get(path, session string, out any) error {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.AddCookie(&http.Cookie{Name: "iamnim_session", Value: session})
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("iamnim %s: status %d", path, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
