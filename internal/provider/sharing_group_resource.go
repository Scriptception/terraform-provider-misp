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
	_ resource.Resource                = (*sharingGroupResource)(nil)
	_ resource.ResourceWithConfigure   = (*sharingGroupResource)(nil)
	_ resource.ResourceWithImportState = (*sharingGroupResource)(nil)
)

// NewSharingGroupResource constructs a misp_sharing_group resource.
func NewSharingGroupResource() resource.Resource {
	return &sharingGroupResource{}
}

type sharingGroupResource struct {
	client *client.Client
}

type sharingGroupResourceModel struct {
	ID            types.String `tfsdk:"id"`
	UUID          types.String `tfsdk:"uuid"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Releasability types.String `tfsdk:"releasability"`
	OrgID         types.String `tfsdk:"org_id"`
	Active        types.Bool   `tfsdk:"active"`
	Local         types.Bool   `tfsdk:"local"`
	Roaming       types.Bool   `tfsdk:"roaming"`
}

func (r *sharingGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sharing_group"
}

func (r *sharingGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A MISP sharing group. Defines which organisations and servers can see data tagged with the group. Org/server membership is managed by future dedicated resources.",
		Attributes: map[string]schema.Attribute{
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
			"name": schema.StringAttribute{
				Description: "Sharing group name.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Free-form description.",
				Optional:    true,
			},
			"releasability": schema.StringAttribute{
				Description: "Human-readable releasability statement (e.g. `For internal use only`).",
				Optional:    true,
				Computed:    true,
			},
			"org_id": schema.StringAttribute{
				Description: "Owning organisation id (defaults to the caller's org).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"active": schema.BoolAttribute{
				Description: "Whether the sharing group is active.",
				Optional:    true,
				Computed:    true,
			},
			"local": schema.BoolAttribute{
				Description: "Local-only: sharing group is not synchronised to other instances.",
				Optional:    true,
				Computed:    true,
			},
			"roaming": schema.BoolAttribute{
				Description: "Roaming: allows member organisations at any instance to receive the data.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *sharingGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *sharingGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sharingGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateSharingGroup(ctx, sgFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Creating MISP sharing group failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, sgToModel(created))...)
}

func (r *sharingGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sharingGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetSharingGroup(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Reading MISP sharing group failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, sgToModel(got))...)
}

func (r *sharingGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan sharingGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state sharingGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateSharingGroup(ctx, state.ID.ValueString(), sgFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Updating MISP sharing group failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, sgToModel(updated))...)
}

func (r *sharingGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sharingGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteSharingGroup(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Deleting MISP sharing group failed", err.Error())
	}
}

func (r *sharingGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func sgFromModel(m sharingGroupResourceModel) client.SharingGroup {
	return client.SharingGroup{
		Name:          m.Name.ValueString(),
		Description:   m.Description.ValueString(),
		Releasability: m.Releasability.ValueString(),
		Active:        m.Active.ValueBool(),
		Local:         m.Local.ValueBool(),
		Roaming:       m.Roaming.ValueBool(),
	}
}

func sgToModel(sg *client.SharingGroup) sharingGroupResourceModel {
	m := sharingGroupResourceModel{
		ID:      types.StringValue(sg.ID),
		UUID:    types.StringValue(sg.UUID),
		Name:    types.StringValue(sg.Name),
		OrgID:   types.StringValue(sg.OrgID),
		Active:  types.BoolValue(sg.Active),
		Local:   types.BoolValue(sg.Local),
		Roaming: types.BoolValue(sg.Roaming),
	}
	m.Description = stringOrNull(sg.Description)
	m.Releasability = stringOrNull(sg.Releasability)
	return m
}
