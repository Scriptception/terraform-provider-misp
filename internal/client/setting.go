package client

import (
	"context"
	"fmt"
)

// Setting mirrors a MISP server-setting object as returned by
// GET /servers/getSetting/{name}. The response is flat (no envelope).
//
// The value field is polymorphic in MISP (string, boolean, or numeric),
// so FlexString absorbs all three JSON representations into a Go string.
type Setting struct {
	Name        string     `json:"name"`
	Value       FlexString `json:"value"`
	Type        string     `json:"type"`
	Description string     `json:"description"`
	Level       int        `json:"level"`
}

// GetSetting fetches a single MISP server setting by its dotted name
// (e.g. "MISP.baseurl"). Returns a NotFound APIError for unknown names.
func (c *Client) GetSetting(ctx context.Context, name string) (*Setting, error) {
	var out Setting
	if err := c.do(ctx, "GET", fmt.Sprintf("/servers/getSetting/%s", name), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateSetting edits a MISP server setting. The value is always sent as a
// JSON string — MISP coerces it to the appropriate type internally.
// No setting data is returned; callers should call GetSetting afterwards to
// capture the authoritative state.
func (c *Client) UpdateSetting(ctx context.Context, name, value string) error {
	body := struct {
		Value string `json:"value"`
	}{Value: value}
	return c.do(ctx, "POST", fmt.Sprintf("/servers/serverSettingsEdit/%s", name), body, nil)
}
