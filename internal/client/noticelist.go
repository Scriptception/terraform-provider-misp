package client

import (
	"context"
	"fmt"
)

// Noticelist mirrors MISP's Noticelist object. Noticelists are bundled with
// MISP; they can't be created from Terraform — only enabled/disabled.
type Noticelist struct {
	ID               string   `json:"id,omitempty"`
	Name             string   `json:"name"`
	ExpandedName     string   `json:"expanded_name,omitempty"`
	Ref              []string `json:"ref,omitempty"`
	GeographicalArea []string `json:"geographical_area,omitempty"`
	Version          string   `json:"version,omitempty"`
	Enabled          FlexBool `json:"enabled"`
}

type noticelistViewEnvelope struct {
	Noticelist Noticelist `json:"Noticelist"`
}

// GetNoticelist fetches a noticelist by id.
func (c *Client) GetNoticelist(ctx context.Context, id string) (*Noticelist, error) {
	var out noticelistViewEnvelope
	if err := c.do(ctx, "GET", fmt.Sprintf("/noticelists/view/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out.Noticelist, nil
}

// ListNoticelists returns all noticelists known to the instance. Used to
// resolve a name to a numeric id, since MISP has no view-by-name endpoint.
// The index response is a flat array: [{"Noticelist": {...}}, ...].
// This is different from warninglists, which are double-wrapped.
func (c *Client) ListNoticelists(ctx context.Context) ([]Noticelist, error) {
	var raw []struct {
		Noticelist Noticelist `json:"Noticelist"`
	}
	if err := c.do(ctx, "GET", "/noticelists", nil, &raw); err != nil {
		return nil, err
	}
	out := make([]Noticelist, 0, len(raw))
	for _, r := range raw {
		out = append(out, r.Noticelist)
	}
	return out, nil
}

// FindNoticelistByName returns the noticelist with the given name, or a
// NotFound APIError if none matches.
func (c *Client) FindNoticelistByName(ctx context.Context, name string) (*Noticelist, error) {
	all, err := c.ListNoticelists(ctx)
	if err != nil {
		return nil, err
	}
	for i := range all {
		if all[i].Name == name {
			return &all[i], nil
		}
	}
	return nil, &APIError{StatusCode: 404, Method: "GET", Path: "/noticelists (filter)", Body: fmt.Sprintf("no noticelist with name %q", name)}
}

// setNoticelistEnabled ensures the noticelist is in the desired enabled state.
// Because MISP's toggleEnable is a pure toggle (no body; flips whatever the
// current state is), we must read first and skip the POST if already correct.
func (c *Client) setNoticelistEnabled(ctx context.Context, id string, target bool) error {
	cur, err := c.GetNoticelist(ctx, id)
	if err != nil {
		return err
	}
	if bool(cur.Enabled) == target {
		return nil
	}
	return c.do(ctx, "POST", fmt.Sprintf("/noticelists/toggleEnable/%s", id), nil, nil)
}

// EnableNoticelist sets enabled=true (idempotent: reads current state first).
func (c *Client) EnableNoticelist(ctx context.Context, id string) error {
	return c.setNoticelistEnabled(ctx, id, true)
}

// DisableNoticelist sets enabled=false (idempotent: reads current state first).
func (c *Client) DisableNoticelist(ctx context.Context, id string) error {
	return c.setNoticelistEnabled(ctx, id, false)
}
