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
	_ resource.Resource                = (*warninglistResource)(nil)
	_ resource.ResourceWithConfigure   = (*warninglistResource)(nil)
	_ resource.ResourceWithImportState = (*warninglistResource)(nil)
)

// NewWarninglistResource constructs a misp_warninglist resource.
//
// Warninglists ship with MISP — this resource adopts an existing one by name
// and manages only its enabled flag. Destroying the resource disables the
// warninglist (does not remove the warninglist definition itself).
func NewWarninglistResource() resource.Resource {
	return &warninglistResource{}
}

type warninglistResource struct {
	client *client.Client
}

type warninglistResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	Description           types.String `tfsdk:"description"`
	Type                  types.String `tfsdk:"type"`
	Version               types.String `tfsdk:"version"`
	Category              types.String `tfsdk:"category"`
	ValidAttributes       types.String `tfsdk:"valid_attributes"`
	WarninglistEntryCount types.String `tfsdk:"warninglist_entry_count"`
}

func (r *warninglistResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_warninglist"
}

func (r *warninglistResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Enable or disable a MISP warninglist. The warninglist definition must already exist on the instance (warninglists are bundled with MISP). Destroying the resource disables the warninglist — it does not remove the definition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Numeric MISP identifier.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Full warninglist name (e.g. `List of known Akamai IP ranges`). Replaces the resource if changed.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the warninglist is active on this instance.",
				Required:    true,
			},
			"description":             schema.StringAttribute{Computed: true},
			"type":                    schema.StringAttribute{Computed: true},
			"version":                 schema.StringAttribute{Computed: true},
			"category":                schema.StringAttribute{Computed: true},
			"valid_attributes":        schema.StringAttribute{Computed: true},
			"warninglist_entry_count": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *warninglistResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *warninglistResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan warninglistResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, err := r.client.FindWarninglistByName(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Warninglist name not found",
			fmt.Sprintf("no warninglist with name %q on this MISP instance. Warninglists must be bundled on the server; run `Warninglists → Update warninglists` in the UI if this is missing.", plan.Name.ValueString()),
		)
		return
	}

	if err := r.applyEnabled(ctx, found.ID, plan.Enabled.ValueBool()); err != nil {
		resp.Diagnostics.AddError("Toggling warninglist failed", err.Error())
		return
	}

	// Re-read to capture the authoritative state after toggling.
	got, err := r.client.GetWarninglist(ctx, found.ID)
	if err != nil {
		resp.Diagnostics.AddError("Reading MISP warninglist failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, warninglistToModel(got))...)
}

func (r *warninglistResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state warninglistResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetWarninglist(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Reading MISP warninglist failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, warninglistToModel(got))...)
}

func (r *warninglistResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan warninglistResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state warninglistResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.applyEnabled(ctx, state.ID.ValueString(), plan.Enabled.ValueBool()); err != nil {
		resp.Diagnostics.AddError("Toggling warninglist failed", err.Error())
		return
	}
	got, err := r.client.GetWarninglist(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Reading MISP warninglist failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, warninglistToModel(got))...)
}

// Delete disables the warninglist without removing its definition. If it's
// already disabled, nothing happens.
func (r *warninglistResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state warninglistResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DisableWarninglist(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Disabling MISP warninglist failed", err.Error())
	}
}

func (r *warninglistResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *warninglistResource) applyEnabled(ctx context.Context, id string, enabled bool) error {
	if enabled {
		return r.client.EnableWarninglist(ctx, id)
	}
	return r.client.DisableWarninglist(ctx, id)
}

func warninglistToModel(w *client.Warninglist) warninglistResourceModel {
	return warninglistResourceModel{
		ID:                    types.StringValue(w.ID),
		Name:                  types.StringValue(w.Name),
		Enabled:               types.BoolValue(w.Enabled),
		Description:           stringOrNull(w.Description),
		Type:                  stringOrNull(w.Type),
		Version:               stringOrNull(w.Version),
		Category:              stringOrNull(w.Category),
		ValidAttributes:       stringOrNull(w.ValidAttributes),
		WarninglistEntryCount: stringOrNull(w.WarninglistEntryCount),
	}
}
