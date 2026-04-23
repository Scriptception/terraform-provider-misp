package client

import (
	"context"
	"fmt"
)

// SharingGroup mirrors MISP's SharingGroup object (core fields only;
// org/server membership is managed via separate endpoints not yet modelled).
type SharingGroup struct {
	ID            string `json:"id,omitempty"`
	UUID          string `json:"uuid,omitempty"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	Releasability string `json:"releasability,omitempty"`
	OrgID         string `json:"org_id,omitempty"`
	Active        bool   `json:"active"`
	Local         bool   `json:"local"`
	Roaming       bool   `json:"roaming"`
}

type sharingGroupEnvelope struct {
	SharingGroup SharingGroup `json:"SharingGroup"`
}

// CreateSharingGroup creates a sharing group.
func (c *Client) CreateSharingGroup(ctx context.Context, sg SharingGroup) (*SharingGroup, error) {
	var out sharingGroupEnvelope
	if err := c.do(ctx, "POST", "/sharing_groups/add", sg, &out); err != nil {
		return nil, err
	}
	return &out.SharingGroup, nil
}

// GetSharingGroup fetches a sharing group by id.
func (c *Client) GetSharingGroup(ctx context.Context, id string) (*SharingGroup, error) {
	var out sharingGroupEnvelope
	if err := c.do(ctx, "GET", fmt.Sprintf("/sharing_groups/view/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out.SharingGroup, nil
}

// UpdateSharingGroup edits a sharing group.
func (c *Client) UpdateSharingGroup(ctx context.Context, id string, sg SharingGroup) (*SharingGroup, error) {
	var out sharingGroupEnvelope
	if err := c.do(ctx, "POST", fmt.Sprintf("/sharing_groups/edit/%s", id), sg, &out); err != nil {
		return nil, err
	}
	return &out.SharingGroup, nil
}

// DeleteSharingGroup removes a sharing group. MISP's API documents DELETE here
// (unlike organisations/tags which accept POST).
func (c *Client) DeleteSharingGroup(ctx context.Context, id string) error {
	return c.do(ctx, "DELETE", fmt.Sprintf("/sharing_groups/delete/%s", id), nil, nil)
}
