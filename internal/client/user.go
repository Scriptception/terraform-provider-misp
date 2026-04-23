package client

import (
	"context"
	"fmt"
)

// User mirrors MISP's User object.
//
// NOTE: authkey is returned only on Create (when advanced_authkeys is disabled)
// and kept out of this v0.1 surface to avoid leaking it into tf state. Manage
// API keys via the MISP UI or a future dedicated resource.
type User struct {
	ID            string `json:"id,omitempty"`
	Email         string `json:"email"`
	OrgID         string `json:"org_id,omitempty"`
	RoleID        string `json:"role_id,omitempty"`
	Autoalert     bool   `json:"autoalert"`
	Contactalert  bool   `json:"contactalert"`
	Disabled      bool   `json:"disabled"`
	Termsaccepted bool   `json:"termsaccepted"`
	ChangePw      bool   `json:"change_pw"`
	GPGKey        string `json:"gpgkey,omitempty"`
	CertifPublic  string `json:"certif_public,omitempty"`
	Expiration    string `json:"expiration,omitempty"`
}

type userEnvelope struct {
	User User `json:"User"`
}

// CreateUser creates a new user. MISP may auto-generate an authkey and send a
// welcome email, depending on instance configuration.
func (c *Client) CreateUser(ctx context.Context, u User) (*User, error) {
	var out userEnvelope
	if err := c.do(ctx, "POST", "/admin/users/add", u, &out); err != nil {
		return nil, err
	}
	return &out.User, nil
}

// GetUser fetches a user by id.
func (c *Client) GetUser(ctx context.Context, id string) (*User, error) {
	var out userEnvelope
	if err := c.do(ctx, "GET", fmt.Sprintf("/admin/users/view/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out.User, nil
}

// UpdateUser edits a user. MISP's API documents PUT for this endpoint.
func (c *Client) UpdateUser(ctx context.Context, id string, u User) (*User, error) {
	var out userEnvelope
	if err := c.do(ctx, "PUT", fmt.Sprintf("/admin/users/edit/%s", id), u, &out); err != nil {
		return nil, err
	}
	return &out.User, nil
}

// DeleteUser removes a user. MISP's API documents DELETE for this endpoint.
func (c *Client) DeleteUser(ctx context.Context, id string) error {
	return c.do(ctx, "DELETE", fmt.Sprintf("/admin/users/delete/%s", id), nil, nil)
}
