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
	_ resource.Resource                = (*tagResource)(nil)
	_ resource.ResourceWithConfigure   = (*tagResource)(nil)
	_ resource.ResourceWithImportState = (*tagResource)(nil)
)

// NewTagResource constructs a misp_tag resource.
func NewTagResource() resource.Resource {
	return &tagResource{}
}

type tagResource struct {
	client *client.Client
}

type tagResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Colour         types.String `tfsdk:"colour"`
	Exportable     types.Bool   `tfsdk:"exportable"`
	HideTag        types.Bool   `tfsdk:"hide_tag"`
	NumericalValue types.Int64  `tfsdk:"numerical_value"`
	OrgID          types.String `tfsdk:"org_id"`
	UserID         types.String `tfsdk:"user_id"`
	LocalOnly      types.Bool   `tfsdk:"local_only"`
}

func (r *tagResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

func (r *tagResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A MISP tag. Tags classify events and attributes and drive taxonomy-based sharing.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Numeric MISP identifier.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Tag name (e.g. `tlp:amber`). Must be unique within the instance.",
				Required:    true,
			},
			"colour": schema.StringAttribute{
				Description: "Hex colour used in the MISP UI (e.g. `#FFC000`).",
				Optional:    true,
				Computed:    true,
			},
			"exportable": schema.BoolAttribute{
				Description: "Whether the tag is included when events are exported. Defaults to true.",
				Optional:    true,
				Computed:    true,
			},
			"hide_tag": schema.BoolAttribute{
				Description: "Hide the tag from the UI tag picker.",
				Optional:    true,
				Computed:    true,
			},
			"numerical_value": schema.Int64Attribute{
				Description: "Optional numeric weight used by analytical workflows.",
				Optional:    true,
			},
			"org_id": schema.StringAttribute{
				Description: "Restrict tag use to an organisation (id). Empty means all organisations.",
				Optional:    true,
				Computed:    true,
			},
			"user_id": schema.StringAttribute{
				Description: "Restrict tag use to a specific user id. Empty means all users.",
				Optional:    true,
				Computed:    true,
			},
			"local_only": schema.BoolAttribute{
				Description: "When true, the tag is not synced to other instances.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *tagResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *tagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tagResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateTag(ctx, tagFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Creating MISP tag failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, tagToModel(created))...)
}

func (r *tagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tagResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetTag(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Reading MISP tag failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, tagToModel(got))...)
}

func (r *tagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan tagResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state tagResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateTag(ctx, state.ID.ValueString(), tagFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Updating MISP tag failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, tagToModel(updated))...)
}

func (r *tagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tagResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteTag(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Deleting MISP tag failed", err.Error())
	}
}

func (r *tagResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func tagFromModel(m tagResourceModel) client.Tag {
	t := client.Tag{
		Name:       m.Name.ValueString(),
		Colour:     m.Colour.ValueString(),
		Exportable: m.Exportable.ValueBool(),
		HideTag:    m.HideTag.ValueBool(),
		OrgID:      m.OrgID.ValueString(),
		UserID:     m.UserID.ValueString(),
		LocalOnly:  m.LocalOnly.ValueBool(),
	}
	if !m.NumericalValue.IsNull() && !m.NumericalValue.IsUnknown() {
		v := m.NumericalValue.ValueInt64()
		t.NumericalValue = &v
	}
	return t
}

func tagToModel(t *client.Tag) tagResourceModel {
	m := tagResourceModel{
		ID:         types.StringValue(t.ID),
		Name:       types.StringValue(t.Name),
		Colour:     types.StringValue(t.Colour),
		Exportable: types.BoolValue(t.Exportable),
		HideTag:    types.BoolValue(t.HideTag),
		OrgID:      types.StringValue(t.OrgID),
		UserID:     types.StringValue(t.UserID),
		LocalOnly:  types.BoolValue(t.LocalOnly),
	}
	if t.NumericalValue != nil {
		m.NumericalValue = types.Int64Value(*t.NumericalValue)
	} else {
		m.NumericalValue = types.Int64Null()
	}
	return m
}
