package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Scriptception/terraform-provider-misp/internal/client"
)

var (
	_ resource.Resource                = (*taxonomyResource)(nil)
	_ resource.ResourceWithConfigure   = (*taxonomyResource)(nil)
	_ resource.ResourceWithImportState = (*taxonomyResource)(nil)
)

// NewTaxonomyResource constructs a misp_taxonomy resource.
//
// Taxonomies ship with MISP — this resource adopts an existing one by
// namespace and manages only its enabled flag. Destroying the resource
// disables the taxonomy (does not remove the taxonomy definition itself).
func NewTaxonomyResource() resource.Resource {
	return &taxonomyResource{}
}

type taxonomyResource struct {
	client *client.Client
}

type taxonomyResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Namespace   types.String `tfsdk:"namespace"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Description types.String `tfsdk:"description"`
	Version     types.String `tfsdk:"version"`
	Exclusive   types.Bool   `tfsdk:"exclusive"`
	Required    types.Bool   `tfsdk:"required"`
}

func (r *taxonomyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_taxonomy"
}

func (r *taxonomyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Enable or disable a MISP taxonomy. The taxonomy definition must already exist on the instance (taxonomies are bundled with MISP). Destroying the resource disables the taxonomy — it does not remove the definition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Numeric MISP identifier.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"namespace": schema.StringAttribute{
				Description: "Stable namespace identifier (e.g. `tlp`, `misp`, `cert-xlm`). Replaces the resource if changed.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the taxonomy is active on this instance.",
				Required:    true,
			},
			"description": schema.StringAttribute{Computed: true},
			"version":     schema.StringAttribute{Computed: true},
			"exclusive":   schema.BoolAttribute{Computed: true},
			"required":    schema.BoolAttribute{Computed: true},
		},
	}
}

func (r *taxonomyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *taxonomyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan taxonomyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, err := r.client.FindTaxonomyByNamespace(ctx, plan.Namespace.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Taxonomy namespace not found",
			fmt.Sprintf("no taxonomy with namespace %q on this MISP instance. Taxonomies must be bundled on the server; run `Taxonomies → Update taxonomies` in the UI if this is missing.", plan.Namespace.ValueString()),
		)
		return
	}

	if err := r.applyEnabled(ctx, found.ID, plan.Enabled.ValueBool()); err != nil {
		resp.Diagnostics.AddError("Toggling taxonomy failed", err.Error())
		return
	}

	// Re-read to capture the authoritative state after toggling.
	got, err := r.client.GetTaxonomy(ctx, found.ID)
	if err != nil {
		resp.Diagnostics.AddError("Reading MISP taxonomy failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, taxonomyToModel(got))...)
}

func (r *taxonomyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state taxonomyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetTaxonomy(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Reading MISP taxonomy failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, taxonomyToModel(got))...)
}

func (r *taxonomyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan taxonomyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state taxonomyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.applyEnabled(ctx, state.ID.ValueString(), plan.Enabled.ValueBool()); err != nil {
		resp.Diagnostics.AddError("Toggling taxonomy failed", err.Error())
		return
	}
	got, err := r.client.GetTaxonomy(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Reading MISP taxonomy failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, taxonomyToModel(got))...)
}

// Delete disables the taxonomy without removing its definition. If it's already
// disabled, nothing happens.
func (r *taxonomyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state taxonomyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DisableTaxonomy(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Disabling MISP taxonomy failed", err.Error())
	}
}

func (r *taxonomyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *taxonomyResource) applyEnabled(ctx context.Context, id string, enabled bool) error {
	if enabled {
		return r.client.EnableTaxonomy(ctx, id)
	}
	return r.client.DisableTaxonomy(ctx, id)
}

func taxonomyToModel(t *client.Taxonomy) taxonomyResourceModel {
	return taxonomyResourceModel{
		ID:          types.StringValue(t.ID),
		Namespace:   types.StringValue(t.Namespace),
		Enabled:     types.BoolValue(t.Enabled),
		Description: stringOrNull(t.Description),
		Version:     stringOrNull(t.Version),
		Exclusive:   types.BoolValue(t.Exclusive),
		Required:    types.BoolValue(t.Required),
	}
}
