package client

// Package client — Server methods.
//
// # Deferred fields
//
//   - pull_rules / push_rules: stringified JSON blobs; need special
//     handling to round-trip through Terraform without spurious diffs.
//   - cert_file / client_cert_file: base64-encoded PEM; marked Sensitive
//     in the API but omitted from GET responses, requiring write-only
//     treatment similar to authkey.
//   - lastpulledid / lastpushedid: operational counters maintained by MISP
//     itself; read-only and not meaningful to declare in config.

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// Server mirrors MISP's Server object (core fields only; see file-level
// comment for fields deferred to a future version).
//
// Organization is a derived field populated by ListServers from the
// RemoteOrg sub-object in the index response; it is not sent in write
// requests (omitempty ensures it stays out of JSON bodies).
type Server struct {
	ID                  string `json:"id,omitempty"`
	Name                string `json:"name"`
	URL                 string `json:"url"`
	Authkey             string `json:"authkey,omitempty"`
	RemoteOrgID         string `json:"remote_org_id"`
	Push                bool   `json:"push"`
	Pull                bool   `json:"pull"`
	PushSightings       bool   `json:"push_sightings"`
	PushGalaxyClusters  bool   `json:"push_galaxy_clusters"`
	PullGalaxyClusters  bool   `json:"pull_galaxy_clusters"`
	SelfSigned          bool   `json:"self_signed"`
	SkipProxy           bool   `json:"skip_proxy"`
	CachingEnabled      bool   `json:"caching_enabled"`
	UnpublishEvent      bool   `json:"unpublish_event"`
	PublishWithoutEmail bool   `json:"publish_without_email"`
	Internal            bool   `json:"internal"`

	// Organization is populated from the RemoteOrg.name field in the index
	// response and is not part of the MISP Server JSON object itself.
	Organization string `json:"-"`
}

// serverEnvelope wraps a Server in MISP's standard `{"Server":{...}}` envelope.
type serverEnvelope struct {
	Server Server `json:"Server"`
}

// serverIndexEntry is the richer structure MISP returns for each element of
// GET /servers/index.  The embedded RemoteOrg carries the remote organisation
// name that MISP learns after the first sync.
type serverIndexEntry struct {
	Server    Server       `json:"Server"`
	RemoteOrg remoteOrgRef `json:"RemoteOrg"`
}

type remoteOrgRef struct {
	Name string `json:"name"`
}

// CreateServer posts the server config. MISP 2.5.x returns 500 from
// /servers/add even when the server is created successfully, so we tolerate
// 5xx here and verify by looking the server up by name in the index. If a
// future MISP version fixes the 500, the happy path returns first and the
// workaround never fires.
func (c *Client) CreateServer(ctx context.Context, s Server) (*Server, error) {
	var out serverEnvelope
	err := c.do(ctx, "POST", "/servers/add", serverEnvelope{Server: s}, &out)
	if err == nil && out.Server.ID != "" {
		return &out.Server, nil
	}
	// 500-tolerance fallback: the resource may have been created despite the
	// error response.
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode < 500 {
		return nil, err // real error, bubble up
	}
	// Look up by name in the index.
	all, listErr := c.ListServers(ctx)
	if listErr != nil {
		return nil, fmt.Errorf("misp: /servers/add returned 500 and list fallback failed: %w (original: %v)", listErr, err)
	}
	for i := range all {
		if all[i].Name == s.Name {
			return &all[i], nil
		}
	}
	return nil, err // nothing found, original error stands
}

// GetServer fetches a server by id. Because MISP offers no /servers/view/{id}
// endpoint, this filters the full index returned by /servers/index.  A
// synthesised 404 APIError is returned when the id is not present so that
// IsNotFound keeps working for the resource's Read method.
func (c *Client) GetServer(ctx context.Context, id string) (*Server, error) {
	all, err := c.ListServers(ctx)
	if err != nil {
		return nil, err
	}
	for i := range all {
		if all[i].ID == id {
			return &all[i], nil
		}
	}
	return nil, &APIError{
		StatusCode: http.StatusNotFound,
		Method:     "GET",
		Path:       fmt.Sprintf("/servers/index (filtering for id=%s)", id),
		Body:       "not found",
	}
}

// UpdateServer edits a server.  MISP 2.5.x suffers the same 500-on-success
// bug as /servers/add, so we apply the same fallback: if a 5xx is received,
// re-fetch the server by id and treat the current state as the result.
func (c *Client) UpdateServer(ctx context.Context, id string, s Server) (*Server, error) {
	var out serverEnvelope
	err := c.do(ctx, "POST", fmt.Sprintf("/servers/edit/%s", id), serverEnvelope{Server: s}, &out)
	if err == nil && out.Server.ID != "" {
		return &out.Server, nil
	}
	// 500-tolerance fallback: the update may have been applied despite the
	// error response.
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode < 500 {
		return nil, err // real error, bubble up
	}
	// Re-fetch by id so the caller gets fresh state.
	got, fetchErr := c.GetServer(ctx, id)
	if fetchErr != nil {
		return nil, fmt.Errorf("misp: /servers/edit/%s returned 500 and re-fetch failed: %w (original: %v)", id, fetchErr, err)
	}
	return got, nil
}

// DeleteServer removes a server. Unlike add/edit, /servers/delete returns a
// normal 200 response and does not suffer from the 500 quirk.
func (c *Client) DeleteServer(ctx context.Context, id string) error {
	return c.do(ctx, "POST", fmt.Sprintf("/servers/delete/%s", id), nil, nil)
}

// ListServers returns all servers visible to the authenticated user.
// The index response is an array of rich objects; we return just the
// Server sub-objects so callers don't have to know about the envelope.
func (c *Client) ListServers(ctx context.Context) ([]Server, error) {
	var entries []serverIndexEntry
	if err := c.do(ctx, "GET", "/servers/index", nil, &entries); err != nil {
		return nil, err
	}
	servers := make([]Server, len(entries))
	for i, e := range entries {
		srv := e.Server
		srv.Organization = e.RemoteOrg.Name
		servers[i] = srv
	}
	return servers, nil
}
