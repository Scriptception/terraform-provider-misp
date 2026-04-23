package client

import (
	"context"
	"fmt"
)

// sharingGroupViewEnvelope decodes the full GET /sharing_groups/view/{id} response,
// which includes the SharingGroupOrg membership array not captured by sharingGroupEnvelope.
type sharingGroupViewEnvelope struct {
	SharingGroup    SharingGroup          `json:"SharingGroup"`
	SharingGroupOrg []sharingGroupOrgItem `json:"SharingGroupOrg"`
}

// sharingGroupOrgItem represents one entry in the SharingGroupOrg array.
type sharingGroupOrgItem struct {
	ID             string `json:"id"`
	SharingGroupID string `json:"sharing_group_id"`
	OrgID          string `json:"org_id"`
	Extend         bool   `json:"extend"`
}

// AddSharingGroupMember adds the given organisation to the sharing group.
func (c *Client) AddSharingGroupMember(ctx context.Context, sgID, orgID string) error {
	return c.do(ctx, "POST", fmt.Sprintf("/sharing_groups/addOrg/%s/%s", sgID, orgID), struct{}{}, nil)
}

// RemoveSharingGroupMember removes the given organisation from the sharing group.
func (c *Client) RemoveSharingGroupMember(ctx context.Context, sgID, orgID string) error {
	return c.do(ctx, "POST", fmt.Sprintf("/sharing_groups/removeOrg/%s/%s", sgID, orgID), struct{}{}, nil)
}

// IsSharingGroupMember checks whether orgID is a member of sgID.
// It returns (true, nil) if the org is present, (false, nil) if the sharing group
// exists but the org is not a member, or (false, err) on any API error.
func (c *Client) IsSharingGroupMember(ctx context.Context, sgID, orgID string) (bool, error) {
	var envelope sharingGroupViewEnvelope
	if err := c.do(ctx, "GET", fmt.Sprintf("/sharing_groups/view/%s", sgID), nil, &envelope); err != nil {
		return false, err
	}
	for _, m := range envelope.SharingGroupOrg {
		if m.OrgID == orgID {
			return true, nil
		}
	}
	return false, nil
}
