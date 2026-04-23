package client

import (
	"context"
	"fmt"
)

// Tag mirrors MISP's Tag object.
type Tag struct {
	ID             string `json:"id,omitempty"`
	Name           string `json:"name"`
	Colour         string `json:"colour,omitempty"`
	Exportable     bool   `json:"exportable"`
	HideTag        bool   `json:"hide_tag"`
	NumericalValue *int64 `json:"numerical_value,omitempty"`
	OrgID          string `json:"org_id,omitempty"`
	UserID         string `json:"user_id,omitempty"`
	LocalOnly      bool   `json:"local_only"`
}

type tagEnvelope struct {
	Tag Tag `json:"Tag"`
}

// CreateTag creates a tag.
func (c *Client) CreateTag(ctx context.Context, t Tag) (*Tag, error) {
	var out tagEnvelope
	if err := c.do(ctx, "POST", "/tags/add", t, &out); err != nil {
		return nil, err
	}
	return &out.Tag, nil
}

// GetTag fetches a tag by numeric id.
//
// Note: MISP's /tags/view endpoint returns a bare Tag object rather than the
// {"Tag": {...}} envelope used by /tags/add and /tags/edit. We therefore decode
// directly into Tag here.
func (c *Client) GetTag(ctx context.Context, id string) (*Tag, error) {
	var out Tag
	if err := c.do(ctx, "GET", fmt.Sprintf("/tags/view/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateTag edits an existing tag.
func (c *Client) UpdateTag(ctx context.Context, id string, t Tag) (*Tag, error) {
	var out tagEnvelope
	if err := c.do(ctx, "POST", fmt.Sprintf("/tags/edit/%s", id), t, &out); err != nil {
		return nil, err
	}
	return &out.Tag, nil
}

// DeleteTag removes a tag.
func (c *Client) DeleteTag(ctx context.Context, id string) error {
	return c.do(ctx, "POST", fmt.Sprintf("/tags/delete/%s", id), nil, nil)
}
