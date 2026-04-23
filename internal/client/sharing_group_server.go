package client

import (
	"context"
	"fmt"
)

// sharingGroupServerViewEnvelope decodes the GET /sharing_groups/view/{id} response
// fields needed to determine server membership. Kept self-contained so this file does
// not depend on the unexported types in sharing_group_member.go.
type sharingGroupServerViewEnvelope struct {
	SharingGroup       SharingGroup              `json:"SharingGroup"`
	SharingGroupServer []sharingGroupServerItem  `json:"SharingGroupServer"`
}

// sharingGroupServerItem represents one entry in the SharingGroupServer array.
type sharingGroupServerItem struct {
	ID             string `json:"id"`
	SharingGroupID string `json:"sharing_group_id"`
	ServerID       string `json:"server_id"`
	AllOrgs        bool   `json:"all_orgs"`
}

// AddSharingGroupServer adds a server to a sharing group.
func (c *Client) AddSharingGroupServer(ctx context.Context, sgID, serverID string) error {
	return c.do(ctx, "POST", fmt.Sprintf("/sharing_groups/addServer/%s/%s", sgID, serverID), struct{}{}, nil)
}

// RemoveSharingGroupServer removes a server from a sharing group.
// NOTE: The OpenAPI spec documents this endpoint as taking (sharingGroupServerId, serverId),
// but the real endpoint expects (sharingGroupId, serverId) — symmetric with addServer.
// Verified empirically on MISP 2.5.36: calling removeServer with the junction row id
// returns "Invalid sharing group or no editing rights."
func (c *Client) RemoveSharingGroupServer(ctx context.Context, sgID, serverID string) error {
	return c.do(ctx, "POST", fmt.Sprintf("/sharing_groups/removeServer/%s/%s", sgID, serverID), struct{}{}, nil)
}

// IsSharingGroupServer returns whether serverID is a member of the sharing group sgID.
// It scans the SharingGroupServer array in GET /sharing_groups/view/{sgId}.
// Returns (true, nil) if present, (false, nil) if the SG exists but the server is not
// listed, or (false, err) on any API error.
func (c *Client) IsSharingGroupServer(ctx context.Context, sgID, serverID string) (bool, error) {
	var envelope sharingGroupServerViewEnvelope
	if err := c.do(ctx, "GET", fmt.Sprintf("/sharing_groups/view/%s", sgID), nil, &envelope); err != nil {
		return false, err
	}
	for _, s := range envelope.SharingGroupServer {
		if s.ServerID == serverID {
			return true, nil
		}
	}
	return false, nil
}
