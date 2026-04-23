package client

import (
	"context"
	"fmt"
)

// Organisation mirrors MISP's Organisation object.
type Organisation struct {
	ID                 string     `json:"id,omitempty"`
	UUID               string     `json:"uuid,omitempty"`
	Name               string     `json:"name"`
	Description        string     `json:"description,omitempty"`
	Type               string     `json:"type,omitempty"`
	Nationality        string     `json:"nationality,omitempty"`
	Sector             string     `json:"sector,omitempty"`
	Contacts           string     `json:"contacts,omitempty"`
	Local              bool       `json:"local"`
	// No omitempty: an empty DomainList must serialize as `[]` on edit to
	// actually clear MISP's stored value. With omitempty, Go would drop the
	// field entirely and MISP would treat it as "no change" — meaning users
	// couldn't empty the domain list once set.
	RestrictedToDomain DomainList `json:"restricted_to_domain"`
	LandingPage        string     `json:"landingpage,omitempty"`
}

type orgEnvelope struct {
	Organisation Organisation `json:"Organisation"`
}

// CreateOrganisation creates an organisation. MISP's admin API returns the created object.
func (c *Client) CreateOrganisation(ctx context.Context, o Organisation) (*Organisation, error) {
	var out orgEnvelope
	if err := c.do(ctx, "POST", "/admin/organisations/add", orgEnvelope{Organisation: o}, &out); err != nil {
		return nil, err
	}
	return &out.Organisation, nil
}

// GetOrganisation fetches a single organisation by id or uuid.
func (c *Client) GetOrganisation(ctx context.Context, id string) (*Organisation, error) {
	var out orgEnvelope
	if err := c.do(ctx, "GET", fmt.Sprintf("/organisations/view/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out.Organisation, nil
}

// UpdateOrganisation edits an existing organisation.
func (c *Client) UpdateOrganisation(ctx context.Context, id string, o Organisation) (*Organisation, error) {
	var out orgEnvelope
	if err := c.do(ctx, "POST", fmt.Sprintf("/admin/organisations/edit/%s", id), orgEnvelope{Organisation: o}, &out); err != nil {
		return nil, err
	}
	return &out.Organisation, nil
}

// DeleteOrganisation removes an organisation.
func (c *Client) DeleteOrganisation(ctx context.Context, id string) error {
	return c.do(ctx, "POST", fmt.Sprintf("/admin/organisations/delete/%s", id), nil, nil)
}
