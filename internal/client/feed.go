package client

import (
	"context"
	"fmt"
)

// Feed mirrors MISP's Feed object (core fields only; operational fields such
// as event_id, settings, and cache_timestamp are excluded from v0.2).
type Feed struct {
	ID             string `json:"id,omitempty"`
	Name           string `json:"name"`
	Provider       string `json:"provider"`
	URL            string `json:"url"`
	SourceFormat   string `json:"source_format,omitempty"`
	Enabled        bool   `json:"enabled"`
	Distribution   string `json:"distribution,omitempty"`
	SharingGroupID string `json:"sharing_group_id,omitempty"`
	TagID          string `json:"tag_id,omitempty"`
	OrgcID         string `json:"orgc_id,omitempty"`
	FixedEvent     bool   `json:"fixed_event"`
	DeltaMerge     bool   `json:"delta_merge"`
	Publish        bool   `json:"publish"`
	OverrideIDs    bool   `json:"override_ids"`
	CachingEnabled bool   `json:"caching_enabled"`
	ForceToIDs     bool   `json:"force_to_ids"`
	LookupVisible  bool   `json:"lookup_visible"`
	InputSource    string `json:"input_source,omitempty"`
	Rules          string `json:"rules,omitempty"`
}

type feedEnvelope struct {
	Feed Feed `json:"Feed"`
}

// CreateFeed creates a feed.
func (c *Client) CreateFeed(ctx context.Context, f Feed) (*Feed, error) {
	var out feedEnvelope
	if err := c.do(ctx, "POST", "/feeds/add", feedEnvelope{Feed: f}, &out); err != nil {
		return nil, err
	}
	return &out.Feed, nil
}

// GetFeed fetches a feed by id.
func (c *Client) GetFeed(ctx context.Context, id string) (*Feed, error) {
	var out feedEnvelope
	if err := c.do(ctx, "GET", fmt.Sprintf("/feeds/view/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out.Feed, nil
}

// UpdateFeed edits a feed.
func (c *Client) UpdateFeed(ctx context.Context, id string, f Feed) (*Feed, error) {
	var out feedEnvelope
	if err := c.do(ctx, "POST", fmt.Sprintf("/feeds/edit/%s", id), feedEnvelope{Feed: f}, &out); err != nil {
		return nil, err
	}
	return &out.Feed, nil
}

// DeleteFeed removes a feed. MISP's API uses DELETE here.
func (c *Client) DeleteFeed(ctx context.Context, id string) error {
	return c.do(ctx, "DELETE", fmt.Sprintf("/feeds/delete/%s", id), nil, nil)
}
