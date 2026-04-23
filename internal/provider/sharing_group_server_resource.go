package provider

// misp_sharing_group_server — junction resource linking a remote MISP server to a sharing group.
//
// Known limitations:
//   - The `all_orgs` flag (include every organisation known to the remote server in the
//     sharing group) is not exposed as an attribute. The addServer endpoint does not accept
//     it at creation time, and updating it post-creation requires patching the parent SG
//     with a full SharingGroupServer array. Use the MISP UI to toggle all_orgs.
//   - server_id="0" is MISP's reserved "this instance" entry, auto-managed by MISP when
//     the sharing group is created. Attempting to manage it via Terraform is rejected.

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
	_ resource.Resource                = (*sharingGroupServerResource)(nil)
	_ resource.ResourceWithConfigure   = (*sharingGroupServerResource)(nil)
	_ resource.ResourceWithImportState = (*sharingGroupServerResource)(nil)
)

// NewSharingGroupServerResource constructs a misp_sharing_group_server resource.
func NewSharingGroupServerResource() resource.Resource {
	return &sharingGroupServerResource{}
}

type sharingGroupServerResource struct {
	client *client.Client
}

type sharingGroupServerResourceModel struct {
	ID             types.String `tfsdk:"id"`
	SharingGroupID types.String `tfsdk:"sharing_group_id"`
	ServerID       types.String `tfsdk:"server_id"`
}

func (r *sharingGroupServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sharing_group_server"
}

func (r *sharingGroupServerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Adds a remote MISP server to a sharing group. " +
			"All attributes require replacement — destroying and re-creating is the only way to change membership. " +
			"The `all_orgs` flag is not supported; use the MISP UI to set it. " +
			"server_id=\"0\" (MISP's auto-managed local-instance entry) is reserved and cannot be managed via Terraform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite identifier in the form `<sharing_group_id>:<server_id>`. " +
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
			"server_id": schema.StringAttribute{
				Description: "Numeric MISP identifier of the remote server to add. " +
					"Value \"0\" is reserved for MISP's auto-managed local-instance entry and is rejected.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *sharingGroupServerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *sharingGroupServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sharingGroupServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sgID := plan.SharingGroupID.ValueString()
	serverID := plan.ServerID.ValueString()

	if serverID == "0" {
		resp.Diagnostics.AddError(
			"Reserved server_id",
			"server_id=\"0\" is MISP's auto-managed local-instance entry and cannot be added or removed via Terraform. "+
				"Use a non-zero server_id referencing a remote MISP server.",
		)
		return
	}

	if err := r.client.AddSharingGroupServer(ctx, sgID, serverID); err != nil {
		resp.Diagnostics.AddError("Adding server to sharing group failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, sharingGroupServerResourceModel{
		ID:             types.StringValue(fmt.Sprintf("%s:%s", sgID, serverID)),
		SharingGroupID: types.StringValue(sgID),
		ServerID:       types.StringValue(serverID),
	})...)
}

func (r *sharingGroupServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sharingGroupServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sgID := state.SharingGroupID.ValueString()
	serverID := state.ServerID.ValueString()

	isMember, err := r.client.IsSharingGroupServer(ctx, sgID, serverID)
	if err != nil {
		if client.IsNotFound(err) {
			// The sharing group itself no longer exists.
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Reading sharing group server membership failed", err.Error())
		return
	}

	if !isMember {
		// Server is no longer in the sharing group — drift detected.
		resp.State.RemoveResource(ctx)
		return
	}

	// Re-set state to keep it consistent (id may not have been set on older imports).
	resp.Diagnostics.Append(resp.State.Set(ctx, sharingGroupServerResourceModel{
		ID:             types.StringValue(fmt.Sprintf("%s:%s", sgID, serverID)),
		SharingGroupID: types.StringValue(sgID),
		ServerID:       types.StringValue(serverID),
	})...)
}

// Update should never be called because all attributes are RequiresReplace.
// Guard defensively in case the framework calls it anyway.
func (r *sharingGroupServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes are RequiresReplace; Terraform should destroy+recreate instead.
	// Re-apply the plan state so the provider does not diverge if called unexpectedly.
	var plan sharingGroupServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *sharingGroupServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sharingGroupServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RemoveSharingGroupServer(ctx, state.SharingGroupID.ValueString(), state.ServerID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Removing server from sharing group failed", err.Error())
	}
}

func (r *sharingGroupServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import id",
			"Expected format: <sharing_group_id>:<server_id> (e.g. \"3:2\")",
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("sharing_group_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("server_id"), parts[1])...)
}
