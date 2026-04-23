package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Scriptception/terraform-provider-misp/internal/client"
)

var (
	_ resource.Resource                = (*galaxyClusterResource)(nil)
	_ resource.ResourceWithConfigure   = (*galaxyClusterResource)(nil)
	_ resource.ResourceWithImportState = (*galaxyClusterResource)(nil)
)

// elementAttrTypes is the attribute type map for a single GalaxyElement object.
// Defined once and reused to build types.ObjectType and types.Object values.
var elementAttrTypes = map[string]attr.Type{
	"key":   types.StringType,
	"value": types.StringType,
}

// NewGalaxyClusterResource constructs a misp_galaxy_cluster resource.
func NewGalaxyClusterResource() resource.Resource {
	return &galaxyClusterResource{}
}

type galaxyClusterResource struct {
	client *client.Client
}

// galaxyClusterResourceModel is the Terraform state/plan model.
type galaxyClusterResourceModel struct {
	// User-settable
	GalaxyID       types.String `tfsdk:"galaxy_id"`
	Value          types.String `tfsdk:"value"`
	Description    types.String `tfsdk:"description"`
	Source         types.String `tfsdk:"source"`
	Authors        types.List   `tfsdk:"authors"`
	Distribution   types.String `tfsdk:"distribution"`
	SharingGroupID types.String `tfsdk:"sharing_group_id"`
	Elements       types.List   `tfsdk:"elements"` // list of {key, value} objects

	// Computed-only
	ID             types.String `tfsdk:"id"`
	UUID           types.String `tfsdk:"uuid"`
	CollectionUUID types.String `tfsdk:"collection_uuid"`
	Type           types.String `tfsdk:"type"`
	TagName        types.String `tfsdk:"tag_name"`
	Version        types.String `tfsdk:"version"`
	OrgID          types.String `tfsdk:"org_id"`
	OrgcID         types.String `tfsdk:"orgc_id"`
	Locked         types.Bool   `tfsdk:"locked"`
	ExtendsUUID    types.String `tfsdk:"extends_uuid"`
	ExtendsVersion types.String `tfsdk:"extends_version"`
	Published      types.Bool   `tfsdk:"published"`
}

func (r *galaxyClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_galaxy_cluster"
}

func (r *galaxyClusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages a **custom** MISP galaxy cluster.

> **Important — bundled clusters cannot be managed by Terraform.**
> MISP ships thousands of bundled clusters (MITRE ATT&CK, threat-actor catalogs, etc.)
> marked with ` + "`default: true`" + `. This resource only manages clusters you create;
> attempting to import a bundled cluster will return an error.
> Use the MISP UI or PyMISP to browse bundled cluster data.

> **Publishing note:** The ` + "`published`" + ` attribute is read-only. Terraform reports
> the current publication state but does not control it. Use the MISP UI or PyMISP to
> publish or unpublish a cluster.`,

		Attributes: map[string]schema.Attribute{
			// ── User-settable ────────────────────────────────────────────
			"galaxy_id": schema.StringAttribute{
				Description: "Numeric id of the parent galaxy. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Description: "Cluster name/value. Must be unique within the galaxy.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Free-form description of the cluster.",
				Optional:    true,
				Computed:    true,
			},
			"source": schema.StringAttribute{
				Description: "Source reference (URL or label) for this cluster.",
				Optional:    true,
				Computed:    true,
			},
			"authors": schema.ListAttribute{
				Description: "List of author names for this cluster.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"distribution": schema.StringAttribute{
				Description: `Numeric distribution level: "0" = Your organisation only, "1" = This community only (default), "2" = Connected communities, "3" = All communities, "4" = Sharing group.`,
				Optional:    true,
				Computed:    true,
			},
			"sharing_group_id": schema.StringAttribute{
				Description: `Sharing group id. Only used when distribution is "4".`,
				Optional:    true,
				Computed:    true,
			},
			"elements": schema.ListNestedAttribute{
				Description: "Key-value metadata for this cluster. Multiple entries may share the same key (e.g. for `refs` or `synonyms`).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key":   schema.StringAttribute{Required: true},
						"value": schema.StringAttribute{Required: true},
					},
				},
			},

			// ── Computed-only ────────────────────────────────────────────
			"id": schema.StringAttribute{
				Description: "Numeric MISP identifier.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uuid": schema.StringAttribute{
				Description: "Stable UUID assigned by MISP.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"collection_uuid": schema.StringAttribute{
				Description: "Collection UUID assigned by MISP.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Description: "Cluster type, derived from the parent galaxy (e.g. `360net-threat-actor`). Set by MISP; cannot be configured.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tag_name": schema.StringAttribute{
				Description: "Tag name MISP derives for this cluster (e.g. `misp-galaxy:type=\"uuid\"`).",
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "Version string set by MISP (unix timestamp of last modification).",
				Computed:    true,
			},
			"org_id": schema.StringAttribute{
				Description: "Owning organisation id (set from the creating user's organisation).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"orgc_id": schema.StringAttribute{
				Description: "Creator organisation id.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"locked": schema.BoolAttribute{
				Description: "Whether the cluster is locked (MISP-managed).",
				Computed:    true,
			},
			"extends_uuid": schema.StringAttribute{
				Description: "UUID of a cluster this one extends, if any.",
				Computed:    true,
			},
			"extends_version": schema.StringAttribute{
				Description: "Version of the extended cluster, if any.",
				Computed:    true,
			},
			"published": schema.BoolAttribute{
				Description: "Whether the cluster has been published. Read-only — Terraform reports current state only. Use the MISP UI or PyMISP to publish or unpublish.",
				Computed:    true,
			},
		},
	}
}

func (r *galaxyClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *galaxyClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan galaxyClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	gc, diags := gcFromModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateGalaxyCluster(ctx, plan.GalaxyID.ValueString(), gc)
	if err != nil {
		resp.Diagnostics.AddError("Creating MISP galaxy cluster failed", err.Error())
		return
	}

	// Re-read to pick up MISP-computed fields (tag_name, version, etc.).
	got, err := r.client.GetGalaxyCluster(ctx, created.ID)
	if err != nil {
		resp.Diagnostics.AddError("Reading MISP galaxy cluster after create failed", err.Error())
		return
	}

	m, diags := gcToModel(ctx, got)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, m)...)
}

func (r *galaxyClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state galaxyClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	got, err := r.client.GetGalaxyCluster(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Reading MISP galaxy cluster failed", err.Error())
		return
	}

	// Soft-deleted clusters are still returned by the API with deleted=true;
	// treat them as absent so Terraform removes them from state.
	if got.Deleted {
		resp.State.RemoveResource(ctx)
		return
	}

	m, diags := gcToModel(ctx, got)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, m)...)
}

func (r *galaxyClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan galaxyClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state galaxyClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	gc, diags := gcFromModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateGalaxyCluster(ctx, state.ID.ValueString(), gc)
	if err != nil {
		resp.Diagnostics.AddError("Updating MISP galaxy cluster failed", err.Error())
		return
	}

	// Re-read to get authoritative state (version ticks, etc.).
	got, err := r.client.GetGalaxyCluster(ctx, updated.ID)
	if err != nil {
		resp.Diagnostics.AddError("Reading MISP galaxy cluster after update failed", err.Error())
		return
	}

	m, diags := gcToModel(ctx, got)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, m)...)
}

func (r *galaxyClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state galaxyClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteGalaxyCluster(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Deleting MISP galaxy cluster failed", err.Error())
	}
}

// ImportState fetches the cluster and rejects bundled (default=true) clusters,
// which cannot be managed by Terraform.
func (r *galaxyClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	got, err := r.client.GetGalaxyCluster(ctx, req.ID)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError("Galaxy cluster not found", fmt.Sprintf("no galaxy cluster with id %q", req.ID))
			return
		}
		resp.Diagnostics.AddError("Reading MISP galaxy cluster failed", err.Error())
		return
	}

	if got.Default {
		resp.Diagnostics.AddError(
			"Cannot import bundled galaxy cluster",
			fmt.Sprintf(
				"Galaxy cluster %q (id %s) is a bundled MISP cluster (default=true). "+
					"Bundled clusters are shipped with MISP and cannot be managed by Terraform. "+
					"Only custom clusters created via this resource may be imported.",
				got.Value, req.ID,
			),
		)
		return
	}

	if got.Deleted {
		resp.Diagnostics.AddError(
			"Cannot import deleted galaxy cluster",
			fmt.Sprintf("Galaxy cluster id %s has been soft-deleted and is no longer manageable.", req.ID),
		)
		return
	}

	m, diags := gcToModel(ctx, got)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, m)...)
}

// ── Conversion helpers ────────────────────────────────────────────────────────

// gcFromModel converts the Terraform plan/state model to a client.GalaxyCluster
// suitable for create/update API calls.
func gcFromModel(ctx context.Context, m galaxyClusterResourceModel) (client.GalaxyCluster, diag.Diagnostics) {
	var diags diag.Diagnostics

	gc := client.GalaxyCluster{
		Value:        m.Value.ValueString(),
		Description:  m.Description.ValueString(),
		Source:       m.Source.ValueString(),
		Distribution: m.Distribution.ValueString(),
	}

	// Authors list
	if !m.Authors.IsNull() && !m.Authors.IsUnknown() {
		var authors []string
		diags.Append(m.Authors.ElementsAs(ctx, &authors, false)...)
		gc.Authors = authors
	}

	// sharing_group_id — only send when non-empty
	if !m.SharingGroupID.IsNull() && !m.SharingGroupID.IsUnknown() && m.SharingGroupID.ValueString() != "" {
		sgid := m.SharingGroupID.ValueString()
		gc.SharingGroupID = &sgid
	}

	// Elements
	if !m.Elements.IsNull() && !m.Elements.IsUnknown() {
		elements, elemDiags := elementsFromList(ctx, m.Elements)
		diags.Append(elemDiags...)
		gc.Elements = elements
	}

	return gc, diags
}

// gcToModel converts a *client.GalaxyCluster to the Terraform model.
func gcToModel(ctx context.Context, gc *client.GalaxyCluster) (galaxyClusterResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	m := galaxyClusterResourceModel{
		ID:             types.StringValue(gc.ID),
		UUID:           types.StringValue(gc.UUID),
		CollectionUUID: types.StringValue(gc.CollectionUUID),
		Type:           types.StringValue(gc.Type),
		GalaxyID:       types.StringValue(gc.GalaxyID),
		Value:          types.StringValue(gc.Value),
		TagName:        types.StringValue(gc.TagName),
		Description:    stringOrNull(gc.Description),
		Source:         stringOrNull(gc.Source),
		Version:        types.StringValue(gc.Version),
		Distribution:   types.StringValue(gc.Distribution),
		OrgID:          types.StringValue(gc.OrgID),
		OrgcID:         types.StringValue(gc.OrgcID),
		Locked:         types.BoolValue(gc.Locked),
		Published:      types.BoolValue(gc.Published),
	}

	// sharing_group_id
	if gc.SharingGroupID != nil && *gc.SharingGroupID != "" {
		m.SharingGroupID = types.StringValue(*gc.SharingGroupID)
	} else {
		m.SharingGroupID = types.StringNull()
	}

	// extends_uuid
	if gc.ExtendsUUID != nil {
		m.ExtendsUUID = types.StringValue(*gc.ExtendsUUID)
	} else {
		m.ExtendsUUID = types.StringNull()
	}

	// extends_version
	if gc.ExtendsVersion != nil {
		m.ExtendsVersion = types.StringValue(*gc.ExtendsVersion)
	} else {
		m.ExtendsVersion = types.StringNull()
	}

	// Authors
	if len(gc.Authors) > 0 {
		authorVals := make([]attr.Value, len(gc.Authors))
		for i, a := range gc.Authors {
			authorVals[i] = types.StringValue(a)
		}
		authorList, authorDiags := types.ListValue(types.StringType, authorVals)
		diags.Append(authorDiags...)
		m.Authors = authorList
	} else {
		m.Authors = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Elements
	elemList, elemDiags := elementsToList(ctx, gc.Elements)
	diags.Append(elemDiags...)
	m.Elements = elemList

	return m, diags
}

// elementsFromList extracts []client.GalaxyElement from a types.List of
// nested objects with "key" and "value" string attributes.
func elementsFromList(ctx context.Context, list types.List) ([]client.GalaxyElement, diag.Diagnostics) {
	var diags diag.Diagnostics

	if list.IsNull() || list.IsUnknown() {
		return nil, diags
	}

	// Extract as []types.Object so we can read each element's attributes.
	var objs []types.Object
	diags.Append(list.ElementsAs(ctx, &objs, false)...)
	if diags.HasError() {
		return nil, diags
	}

	elements := make([]client.GalaxyElement, 0, len(objs))
	for _, obj := range objs {
		attrs := obj.Attributes()
		keyAttr, ok1 := attrs["key"].(types.String)
		valAttr, ok2 := attrs["value"].(types.String)
		if !ok1 || !ok2 {
			diags.AddError("Element attribute type error", "expected string attributes 'key' and 'value' in elements entry")
			continue
		}
		elements = append(elements, client.GalaxyElement{
			Key:   keyAttr.ValueString(),
			Value: valAttr.ValueString(),
		})
	}
	return elements, diags
}

// elementsToList builds a types.List of nested objects from []client.GalaxyElement.
func elementsToList(_ context.Context, elements []client.GalaxyElement) (types.List, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: elementAttrTypes}

	if len(elements) == 0 {
		return types.ListValueMust(elemType, []attr.Value{}), nil
	}

	objVals := make([]attr.Value, 0, len(elements))
	var diags diag.Diagnostics
	for _, e := range elements {
		obj, objDiags := types.ObjectValue(elementAttrTypes, map[string]attr.Value{
			"key":   types.StringValue(e.Key),
			"value": types.StringValue(e.Value),
		})
		diags.Append(objDiags...)
		if !objDiags.HasError() {
			objVals = append(objVals, obj)
		}
	}
	if diags.HasError() {
		return types.ListNull(elemType), diags
	}

	list, listDiags := types.ListValue(elemType, objVals)
	diags.Append(listDiags...)
	return list, diags
}
