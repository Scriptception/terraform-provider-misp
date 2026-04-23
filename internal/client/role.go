package client

import (
	"context"
	"fmt"
	"net/http"
)

// Role mirrors MISP's Role object.
type Role struct {
	ID                   string `json:"id,omitempty"`
	Name                 string `json:"name"`
	Permission           string `json:"permission"`
	PermAdd              FlexBool `json:"perm_add"`
	PermModify           FlexBool `json:"perm_modify"`
	PermModifyOrg        FlexBool `json:"perm_modify_org"`
	PermPublish          FlexBool `json:"perm_publish"`
	PermDelegate         FlexBool `json:"perm_delegate"`
	PermSighting         FlexBool `json:"perm_sighting"`
	PermTagger           FlexBool `json:"perm_tagger"`
	PermTemplate         FlexBool `json:"perm_template"`
	PermSharingGroup     FlexBool `json:"perm_sharing_group"`
	PermTagEditor        FlexBool `json:"perm_tag_editor"`
	PermObjectTemplate   FlexBool `json:"perm_object_template"`
	PermSync             FlexBool `json:"perm_sync"`
	PermAdmin            FlexBool `json:"perm_admin"`
	PermAuth             FlexBool `json:"perm_auth"`
	PermSiteAdmin        FlexBool `json:"perm_site_admin"`
	PermRegexpAccess     FlexBool `json:"perm_regexp_access"`
	PermAudit            FlexBool `json:"perm_audit"`
	PermPublishZmq       FlexBool `json:"perm_publish_zmq"`
	PermPublishKafka     FlexBool `json:"perm_publish_kafka"`
	PermDecaying         FlexBool `json:"perm_decaying"`
	PermGalaxyEditor     FlexBool `json:"perm_galaxy_editor"`
	DefaultRole          FlexBool `json:"default_role"`
	RestrictedToSiteAdmin FlexBool `json:"restricted_to_site_admin"`
	EnforceRateLimit     FlexBool `json:"enforce_rate_limit"`
	MemoryLimit           FlexString `json:"memory_limit,omitempty"`
	MaxExecutionTime      FlexString `json:"max_execution_time,omitempty"`
	RateLimitCount        FlexString `json:"rate_limit_count,omitempty"`
}

type roleEnvelope struct {
	Role Role `json:"Role"`
}

type roleListItem struct {
	Role Role `json:"Role"`
}

// CreateRole creates a new MISP role.
func (c *Client) CreateRole(ctx context.Context, r Role) (*Role, error) {
	body := roleEnvelope{Role: r}
	var out roleEnvelope
	if err := c.do(ctx, "POST", "/admin/roles/add", body, &out); err != nil {
		return nil, err
	}
	return &out.Role, nil
}

// GetRole fetches a role by id. MISP has no single-role endpoint; we list all
// roles and filter. A synthetic 404 APIError is returned when the id is not found.
func (c *Client) GetRole(ctx context.Context, id string) (*Role, error) {
	var list []roleListItem
	if err := c.do(ctx, "GET", "/roles", nil, &list); err != nil {
		return nil, err
	}
	for _, item := range list {
		if item.Role.ID == id {
			r := item.Role
			return &r, nil
		}
	}
	return nil, &APIError{
		StatusCode: http.StatusNotFound,
		Method:     "GET",
		Path:       fmt.Sprintf("/roles/%s", id),
		Body:       "role not found",
	}
}

// UpdateRole edits an existing MISP role.
func (c *Client) UpdateRole(ctx context.Context, id string, r Role) (*Role, error) {
	body := roleEnvelope{Role: r}
	var out roleEnvelope
	if err := c.do(ctx, "POST", fmt.Sprintf("/admin/roles/edit/%s", id), body, &out); err != nil {
		return nil, err
	}
	return &out.Role, nil
}

// DeleteRole removes a MISP role.
func (c *Client) DeleteRole(ctx context.Context, id string) error {
	return c.do(ctx, "POST", fmt.Sprintf("/admin/roles/delete/%s", id), nil, nil)
}
