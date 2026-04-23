package client

import (
	"context"
	"fmt"
)

// Warninglist mirrors MISP's Warninglist object. Warninglists are bundled with
// MISP; they can't be created from Terraform — only enabled/disabled.
type Warninglist struct {
	ID                   string `json:"id,omitempty"`
	Name                 string `json:"name"`
	Type                 string `json:"type,omitempty"`
	Description          string `json:"description,omitempty"`
	Version              string `json:"version,omitempty"`
	Enabled              bool   `json:"enabled"`
	Category             string `json:"category,omitempty"`
	WarninglistEntryCount string `json:"warninglist_entry_count,omitempty"`
	ValidAttributes      string `json:"valid_attributes,omitempty"`
}

type warninglistViewEnvelope struct {
	Warninglist Warninglist `json:"Warninglist"`
}

// GetWarninglist fetches a warninglist by id.
func (c *Client) GetWarninglist(ctx context.Context, id string) (*Warninglist, error) {
	var out warninglistViewEnvelope
	if err := c.do(ctx, "GET", fmt.Sprintf("/warninglists/view/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out.Warninglist, nil
}

// ListWarninglists returns all warninglists known to the instance. Used to
// resolve a name to a numeric id, since MISP has no view-by-name endpoint.
// The index response is double-wrapped: {"Warninglists": [{"Warninglist": {...}}, ...]}.
func (c *Client) ListWarninglists(ctx context.Context) ([]Warninglist, error) {
	var raw struct {
		Warninglists []struct {
			Warninglist Warninglist `json:"Warninglist"`
		} `json:"Warninglists"`
	}
	if err := c.do(ctx, "GET", "/warninglists/index", nil, &raw); err != nil {
		return nil, err
	}
	out := make([]Warninglist, 0, len(raw.Warninglists))
	for _, r := range raw.Warninglists {
		out = append(out, r.Warninglist)
	}
	return out, nil
}

// FindWarninglistByName returns the warninglist with the given name, or a
// NotFound APIError if none matches.
func (c *Client) FindWarninglistByName(ctx context.Context, name string) (*Warninglist, error) {
	all, err := c.ListWarninglists(ctx)
	if err != nil {
		return nil, err
	}
	for i := range all {
		if all[i].Name == name {
			return &all[i], nil
		}
	}
	return nil, &APIError{StatusCode: 404, Method: "GET", Path: "/warninglists (filter)", Body: fmt.Sprintf("no warninglist with name %q", name)}
}

// EnableWarninglist marks a warninglist enabled (idempotent on MISP's side).
func (c *Client) EnableWarninglist(ctx context.Context, id string) error {
	body := map[string]any{"id": id, "enabled": true}
	return c.do(ctx, "POST", "/warninglists/toggleEnable", body, nil)
}

// DisableWarninglist marks a warninglist disabled (idempotent on MISP's side).
func (c *Client) DisableWarninglist(ctx context.Context, id string) error {
	body := map[string]any{"id": id, "enabled": false}
	return c.do(ctx, "POST", "/warninglists/toggleEnable", body, nil)
}
