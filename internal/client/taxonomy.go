package client

import (
	"context"
	"fmt"
)

// Taxonomy mirrors MISP's Taxonomy object. Taxonomies are bundled with MISP;
// they can't be created from Terraform — only enabled/disabled.
type Taxonomy struct {
	ID          string `json:"id,omitempty"`
	Namespace   string `json:"namespace"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
	Enabled     bool   `json:"enabled"`
	Exclusive   bool   `json:"exclusive"`
	Required    bool   `json:"required"`
	Highlighted bool   `json:"highlighted,omitempty"`
}

type taxonomyViewEnvelope struct {
	Taxonomy Taxonomy `json:"Taxonomy"`
}

// GetTaxonomy fetches a taxonomy by id.
func (c *Client) GetTaxonomy(ctx context.Context, id string) (*Taxonomy, error) {
	var out taxonomyViewEnvelope
	if err := c.do(ctx, "GET", fmt.Sprintf("/taxonomies/view/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out.Taxonomy, nil
}

// ListTaxonomies returns all taxonomies known to the instance. Used to resolve
// a namespace to a numeric id, since MISP has no view-by-namespace endpoint.
func (c *Client) ListTaxonomies(ctx context.Context) ([]Taxonomy, error) {
	var raw []struct {
		Taxonomy Taxonomy `json:"Taxonomy"`
	}
	if err := c.do(ctx, "GET", "/taxonomies/index.json", nil, &raw); err != nil {
		return nil, err
	}
	out := make([]Taxonomy, 0, len(raw))
	for _, r := range raw {
		out = append(out, r.Taxonomy)
	}
	return out, nil
}

// FindTaxonomyByNamespace returns the taxonomy with the given namespace, or
// a NotFound APIError if none matches.
func (c *Client) FindTaxonomyByNamespace(ctx context.Context, namespace string) (*Taxonomy, error) {
	all, err := c.ListTaxonomies(ctx)
	if err != nil {
		return nil, err
	}
	for i := range all {
		if all[i].Namespace == namespace {
			return &all[i], nil
		}
	}
	return nil, &APIError{StatusCode: 404, Method: "GET", Path: "/taxonomies (filter)", Body: fmt.Sprintf("no taxonomy with namespace %q", namespace)}
}

// EnableTaxonomy marks a taxonomy enabled (idempotent on MISP's side).
func (c *Client) EnableTaxonomy(ctx context.Context, id string) error {
	return c.do(ctx, "POST", fmt.Sprintf("/taxonomies/enable/%s", id), nil, nil)
}

// DisableTaxonomy marks a taxonomy disabled (idempotent on MISP's side).
func (c *Client) DisableTaxonomy(ctx context.Context, id string) error {
	return c.do(ctx, "POST", fmt.Sprintf("/taxonomies/disable/%s", id), nil, nil)
}
