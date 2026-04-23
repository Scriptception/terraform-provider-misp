package provider

// misp_sharing_group_member — junction resource linking an organisation to a sharing group.
//
// Known limitation: the `extend` flag (allows a member org to re-share the SG) is not
// exposed as an attribute. The MISP addOrg endpoint does not accept it at creation
// time, and patching it post-creation via /sharing_groups/edit requires a complex
// envelope. Use the MISP UI to toggle extend=true on individual memberships.

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Scriptception/terraform-provider-misp/internal/client"
)

var (
	_ resource.Resource                = (*sharingGroupMemberResource)(nil)
	_ resource.ResourceWithConfigure   = (*sharingGroupMemberResource)(nil)
	_ resource.ResourceWithImportState = (*sharingGroupMemberResource)(nil)
)

// NewSharingGroupMemberResource constructs a misp_sharing_group_member resource.
func NewSharingGroupMemberResource() resource.Resource {
	return &sharingGroupMemberResource{}
}

type sharingGroupMemberResource struct {
	client *client.Client
}

type sharingGroupMemberResourceModel struct {
	ID             types.String `tfsdk:"id"`
	SharingGroupID types.String `tfsdk:"sharing_group_id"`
	OrganisationID types.String `tfsdk:"organisation_id"`
}

func (r *sharingGroupMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sharing_group_member"
}

func (r *sharingGroupMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Adds an organisation as a member of a MISP sharing group. " +
			"All attributes require replacement — destroying and re-creating is the only way to change membership. " +
			"The `extend` flag is not supported; use the MISP UI to set it.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite identifier in the form `<sharing_group_id>:<organisation_id>`. " +
					"Used for import; not settable by the user.",
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"sharing_group_id": schema.StringAttribute{
				Description: "Numeric MISP identifier of the sharing group.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"organisation_id": schema.StringAttribute{
				Description: "Numeric MISP identifier of the organisation to add as a member.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *sharingGroupMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("expected *client.Client, got %T", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *sharingGroupMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sharingGroupMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sgID := plan.SharingGroupID.ValueString()
	orgID := plan.OrganisationID.ValueString()

	if err := r.client.AddSharingGroupMember(ctx, sgID, orgID); err != nil {
		resp.Diagnostics.AddError("Adding organisation to sharing group failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, sharingGroupMemberResourceModel{
		ID:             types.StringValue(fmt.Sprintf("%s:%s", sgID, orgID)),
		SharingGroupID: types.StringValue(sgID),
		OrganisationID: types.StringValue(orgID),
	})...)
}

func (r *sharingGroupMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sharingGroupMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sgID := state.SharingGroupID.ValueString()
	orgID := state.OrganisationID.ValueString()

	isMember, err := r.client.IsSharingGroupMember(ctx, sgID, orgID)
	if err != nil {
		if client.IsNotFound(err) {
			// The sharing group itself no longer exists.
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Reading sharing group membership failed", err.Error())
		return
	}

	if !isMember {
		// Org is no longer in the sharing group — drift detected.
		resp.State.RemoveResource(ctx)
		return
	}

	// Re-set state to keep it consistent (id may not have been set on older imports).
	resp.Diagnostics.Append(resp.State.Set(ctx, sharingGroupMemberResourceModel{
		ID:             types.StringValue(fmt.Sprintf("%s:%s", sgID, orgID)),
		SharingGroupID: types.StringValue(sgID),
		OrganisationID: types.StringValue(orgID),
	})...)
}

// Update should never be called because all attributes are RequiresReplace.
// Guard defensively in case the framework calls it anyway.
func (r *sharingGroupMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes are RequiresReplace; Terraform should destroy+recreate instead.
	// Re-apply the plan state so the provider does not diverge if called unexpectedly.
	var plan sharingGroupMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *sharingGroupMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sharingGroupMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RemoveSharingGroupMember(ctx, state.SharingGroupID.ValueString(), state.OrganisationID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Removing organisation from sharing group failed", err.Error())
	}
}

func (r *sharingGroupMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import id",
			"Expected format: <sharing_group_id>:<organisation_id> (e.g. \"3:7\")",
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("sharing_group_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organisation_id"), parts[1])...)
}
