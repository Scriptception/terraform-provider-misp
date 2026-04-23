package client

import (
	"context"
	"fmt"
)

// GalaxyElement is a single key/value metadata entry for a GalaxyCluster.
// MISP allows multiple entries with the same key (e.g. several "refs" or
// "synonyms"). The server-side id and galaxy_cluster_id are not modelled here
// as they are ephemeral artifacts managed by MISP.
type GalaxyElement struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// GalaxyCluster mirrors MISP's GalaxyCluster object.
//
// Only fields needed for create/update are sent in request bodies. All fields
// are decoded on read. The Default field is always false for user-created
// clusters; bundled clusters (default=true) are not managed by Terraform.
type GalaxyCluster struct {
	ID             string          `json:"id,omitempty"`
	UUID           string          `json:"uuid,omitempty"`
	CollectionUUID string          `json:"collection_uuid,omitempty"`
	Type           string          `json:"type,omitempty"`
	Value          string          `json:"value,omitempty"`
	TagName        string          `json:"tag_name,omitempty"`
	Description    string          `json:"description,omitempty"`
	GalaxyID       string          `json:"galaxy_id,omitempty"`
	Source         string          `json:"source,omitempty"`
	Authors        []string        `json:"authors,omitempty"`
	Version        string          `json:"version,omitempty"`
	Distribution   string          `json:"distribution,omitempty"`
	SharingGroupID *string         `json:"sharing_group_id,omitempty"`
	OrgID          string          `json:"org_id,omitempty"`
	OrgcID         string          `json:"orgc_id,omitempty"`
	Default        bool            `json:"default,omitempty"`
	Locked         bool            `json:"locked,omitempty"`
	ExtendsUUID    *string         `json:"extends_uuid,omitempty"`
	ExtendsVersion *string         `json:"extends_version,omitempty"`
	Published      bool            `json:"published,omitempty"`
	Deleted        bool            `json:"deleted,omitempty"`
	Elements       []GalaxyElement `json:"GalaxyElement,omitempty"`
}

// galaxyClusterRequest is the wire body for create and edit calls.
type galaxyClusterRequest struct {
	GalaxyCluster galaxyClusterPayload `json:"GalaxyCluster"`
}

// galaxyClusterPayload contains only the user-settable fields sent in
// create/update requests. It explicitly omits server-managed fields.
type galaxyClusterPayload struct {
	Value          string          `json:"value,omitempty"`
	Description    string          `json:"description,omitempty"`
	Source         string          `json:"source,omitempty"`
	Authors        []string        `json:"authors,omitempty"`
	Distribution   string          `json:"distribution,omitempty"`
	SharingGroupID string          `json:"sharing_group_id,omitempty"`
	Elements       []GalaxyElement `json:"GalaxyElement,omitempty"`
}

type galaxyClusterEnvelope struct {
	GalaxyCluster galaxyClusterRead `json:"GalaxyCluster"`
}

// galaxyClusterRead is the full response shape, including server-set fields.
// We decode into a separate struct so that fields like "default" (which use
// reserved Go identifiers with meaning) are handled cleanly. The decoded
// result is then mapped to GalaxyCluster.
type galaxyClusterRead struct {
	ID             string          `json:"id"`
	UUID           string          `json:"uuid"`
	CollectionUUID string          `json:"collection_uuid"`
	Type           string          `json:"type"`
	Value          string          `json:"value"`
	TagName        string          `json:"tag_name"`
	Description    string          `json:"description"`
	GalaxyID       string          `json:"galaxy_id"`
	Source         string          `json:"source"`
	Authors        []string        `json:"authors"`
	Version        string          `json:"version"`
	Distribution   string          `json:"distribution"`
	SharingGroupID *string         `json:"sharing_group_id"`
	OrgID          string          `json:"org_id"`
	OrgcID         string          `json:"orgc_id"`
	Default        bool            `json:"default"`
	Locked         bool            `json:"locked"`
	ExtendsUUID    *string         `json:"extends_uuid"`
	ExtendsVersion *string         `json:"extends_version"`
	Published      bool            `json:"published"`
	Deleted        bool            `json:"deleted"`
	Elements       []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"GalaxyElement"`
}

func (r galaxyClusterRead) toGalaxyCluster() *GalaxyCluster {
	gc := &GalaxyCluster{
		ID:             r.ID,
		UUID:           r.UUID,
		CollectionUUID: r.CollectionUUID,
		Type:           r.Type,
		Value:          r.Value,
		TagName:        r.TagName,
		Description:    r.Description,
		GalaxyID:       r.GalaxyID,
		Source:         r.Source,
		Authors:        r.Authors,
		Version:        r.Version,
		Distribution:   r.Distribution,
		SharingGroupID: r.SharingGroupID,
		OrgID:          r.OrgID,
		OrgcID:         r.OrgcID,
		Default:        r.Default,
		Locked:         r.Locked,
		ExtendsUUID:    r.ExtendsUUID,
		ExtendsVersion: r.ExtendsVersion,
		Published:      r.Published,
		Deleted:        r.Deleted,
	}
	if len(r.Elements) > 0 {
		gc.Elements = make([]GalaxyElement, len(r.Elements))
		for i, e := range r.Elements {
			gc.Elements[i] = GalaxyElement{Key: e.Key, Value: e.Value}
		}
	}
	return gc
}

// CreateGalaxyCluster creates a new custom galaxy cluster inside the given galaxy.
// The GalaxyCluster.Default field is always set to false by MISP for user-created
// clusters — callers must not set it.
func (c *Client) CreateGalaxyCluster(ctx context.Context, galaxyID string, gc GalaxyCluster) (*GalaxyCluster, error) {
	payload := galaxyClusterPayload{
		Value:       gc.Value,
		Description: gc.Description,
		Source:      gc.Source,
		Authors:     gc.Authors,
		Distribution: gc.Distribution,
		Elements:    gc.Elements,
	}
	if gc.SharingGroupID != nil && *gc.SharingGroupID != "" {
		payload.SharingGroupID = *gc.SharingGroupID
	}
	body := galaxyClusterRequest{GalaxyCluster: payload}

	var out galaxyClusterEnvelope
	if err := c.do(ctx, "POST", fmt.Sprintf("/galaxy_clusters/add/%s", galaxyID), body, &out); err != nil {
		return nil, err
	}
	return out.GalaxyCluster.toGalaxyCluster(), nil
}

// GetGalaxyCluster fetches a galaxy cluster by numeric id.
func (c *Client) GetGalaxyCluster(ctx context.Context, id string) (*GalaxyCluster, error) {
	var out galaxyClusterEnvelope
	if err := c.do(ctx, "GET", fmt.Sprintf("/galaxy_clusters/view/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return out.GalaxyCluster.toGalaxyCluster(), nil
}

// UpdateGalaxyCluster edits an existing galaxy cluster. When GalaxyElement is
// included, MISP replaces the entire element list for the cluster.
func (c *Client) UpdateGalaxyCluster(ctx context.Context, id string, gc GalaxyCluster) (*GalaxyCluster, error) {
	payload := galaxyClusterPayload{
		Value:       gc.Value,
		Description: gc.Description,
		Source:      gc.Source,
		Authors:     gc.Authors,
		Distribution: gc.Distribution,
		Elements:    gc.Elements,
	}
	if gc.SharingGroupID != nil && *gc.SharingGroupID != "" {
		payload.SharingGroupID = *gc.SharingGroupID
	}
	body := galaxyClusterRequest{GalaxyCluster: payload}

	var out galaxyClusterEnvelope
	if err := c.do(ctx, "POST", fmt.Sprintf("/galaxy_clusters/edit/%s", id), body, &out); err != nil {
		return nil, err
	}
	return out.GalaxyCluster.toGalaxyCluster(), nil
}

// DeleteGalaxyCluster soft-deletes a galaxy cluster. MISP marks the record as
// deleted=true rather than removing it; subsequent GET calls return HTTP 200
// with deleted=true, which the provider treats as absent.
func (c *Client) DeleteGalaxyCluster(ctx context.Context, id string) error {
	return c.do(ctx, "POST", fmt.Sprintf("/galaxy_clusters/delete/%s", id), nil, nil)
}
