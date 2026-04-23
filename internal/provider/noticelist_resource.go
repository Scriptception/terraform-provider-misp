package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Scriptception/terraform-provider-misp/internal/client"
)

var (
	_ resource.Resource                = (*noticelistResource)(nil)
	_ resource.ResourceWithConfigure   = (*noticelistResource)(nil)
	_ resource.ResourceWithImportState = (*noticelistResource)(nil)
)

// NewNoticelistResource constructs a misp_noticelist resource.
//
// Noticelists ship with MISP — this resource adopts an existing one by name
// and manages only its enabled flag. Destroying the resource disables the
// noticelist (does not remove the noticelist definition itself).
func NewNoticelistResource() resource.Resource {
	return &noticelistResource{}
}

type noticelistResource struct {
	client *client.Client
}

type noticelistResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	ExpandedName     types.String `tfsdk:"expanded_name"`
	Version          types.String `tfsdk:"version"`
	Ref              types.List   `tfsdk:"ref"`
	GeographicalArea types.List   `tfsdk:"geographical_area"`
}

func (r *noticelistResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_noticelist"
}

func (r *noticelistResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Enable or disable a MISP noticelist. The noticelist definition must already exist on the instance (noticelists are bundled with MISP). Destroying the resource disables the noticelist — it does not remove the definition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Numeric MISP identifier.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Short noticelist name (e.g. `gdpr`). Replaces the resource if changed.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the noticelist is active on this instance.",
				Required:    true,
			},
			"expanded_name": schema.StringAttribute{
				Description: "Human-readable noticelist name.",
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "Noticelist version string.",
				Computed:    true,
			},
			"ref": schema.ListAttribute{
				Description: "Reference URLs for this noticelist.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"geographical_area": schema.ListAttribute{
				Description: "Geographical areas covered by this noticelist.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *noticelistResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *noticelistResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan noticelistResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, err := r.client.FindNoticelistByName(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Noticelist name not found",
			fmt.Sprintf("no noticelist with name %q on this MISP instance. Noticelists must be bundled on the server; check the UI if this is missing.", plan.Name.ValueString()),
		)
		return
	}

	if err := r.applyEnabled(ctx, found.ID, plan.Enabled.ValueBool()); err != nil {
		resp.Diagnostics.AddError("Toggling noticelist failed", err.Error())
		return
	}

	// Re-read to capture the authoritative state after toggling.
	got, err := r.client.GetNoticelist(ctx, found.ID)
	if err != nil {
		resp.Diagnostics.AddError("Reading MISP noticelist failed", err.Error())
		return
	}
	m, diags := noticelistToModel(ctx, got)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, m)...)
}

func (r *noticelistResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state noticelistResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetNoticelist(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Reading MISP noticelist failed", err.Error())
		return
	}
	m, diags := noticelistToModel(ctx, got)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, m)...)
}

func (r *noticelistResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan noticelistResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state noticelistResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.applyEnabled(ctx, state.ID.ValueString(), plan.Enabled.ValueBool()); err != nil {
		resp.Diagnostics.AddError("Toggling noticelist failed", err.Error())
		return
	}
	got, err := r.client.GetNoticelist(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Reading MISP noticelist failed", err.Error())
		return
	}
	m, diags := noticelistToModel(ctx, got)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, m)...)
}

// Delete disables the noticelist without removing its definition. If it's
// already disabled, nothing happens (setNoticelistEnabled is idempotent).
func (r *noticelistResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state noticelistResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DisableNoticelist(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Disabling MISP noticelist failed", err.Error())
	}
}

func (r *noticelistResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *noticelistResource) applyEnabled(ctx context.Context, id string, enabled bool) error {
	if enabled {
		return r.client.EnableNoticelist(ctx, id)
	}
	return r.client.DisableNoticelist(ctx, id)
}

// noticelistToModel converts a client Noticelist to the Terraform resource model.
// It returns diagnostics because building list values can produce warnings/errors.
func noticelistToModel(ctx context.Context, n *client.Noticelist) (noticelistResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	refList, refDiags := types.ListValueFrom(ctx, types.StringType, n.Ref)
	diags.Append(refDiags...)

	geoList, geoDiags := types.ListValueFrom(ctx, types.StringType, n.GeographicalArea)
	diags.Append(geoDiags...)

	m := noticelistResourceModel{
		ID:               types.StringValue(n.ID),
		Name:             types.StringValue(n.Name),
		Enabled:          types.BoolValue(bool(n.Enabled)),
		ExpandedName:     stringOrNull(n.ExpandedName),
		Version:          stringOrNull(n.Version),
		Ref:              refList,
		GeographicalArea: geoList,
	}
	return m, diags
}
