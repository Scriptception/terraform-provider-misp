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
	_ resource.Resource                = (*userResource)(nil)
	_ resource.ResourceWithConfigure   = (*userResource)(nil)
	_ resource.ResourceWithImportState = (*userResource)(nil)
)

// NewUserResource constructs a misp_user resource.
func NewUserResource() resource.Resource {
	return &userResource{}
}

type userResource struct {
	client *client.Client
}

type userResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Email         types.String `tfsdk:"email"`
	OrgID         types.String `tfsdk:"org_id"`
	RoleID        types.String `tfsdk:"role_id"`
	Autoalert     types.Bool   `tfsdk:"autoalert"`
	Contactalert  types.Bool   `tfsdk:"contactalert"`
	Disabled      types.Bool   `tfsdk:"disabled"`
	Termsaccepted types.Bool   `tfsdk:"termsaccepted"`
	ChangePw      types.Bool   `tfsdk:"change_pw"`
	GPGKey        types.String `tfsdk:"gpgkey"`
	CertifPublic  types.String `tfsdk:"certif_public"`
	Expiration    types.String `tfsdk:"expiration"`
}

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A MISP user. Creating a user may cause MISP to send a welcome email; API keys are not exposed here — manage them separately.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Numeric MISP identifier.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				Description: "User email address. Must be unique within the instance.",
				Required:    true,
			},
			"org_id": schema.StringAttribute{
				Description: "Owning organisation id.",
				Required:    true,
			},
			"role_id": schema.StringAttribute{
				Description: "MISP role id (e.g. 1=admin, 3=user, 6=read-only). See `/roles` on your instance.",
				Required:    true,
			},
			"disabled": schema.BoolAttribute{
				Description: "Disable login for this user.",
				Optional:    true,
				Computed:    true,
			},
			"autoalert": schema.BoolAttribute{
				Description: "Receive automatic email alerts for new events.",
				Optional:    true,
				Computed:    true,
			},
			"contactalert": schema.BoolAttribute{
				Description: "Receive contact requests via email.",
				Optional:    true,
				Computed:    true,
			},
			"termsaccepted": schema.BoolAttribute{
				Description: "Whether the user has accepted MISP's terms.",
				Optional:    true,
				Computed:    true,
			},
			"change_pw": schema.BoolAttribute{
				Description: "Require password change on next login.",
				Optional:    true,
				Computed:    true,
			},
			"gpgkey": schema.StringAttribute{
				Description: "Armored GPG public key used to encrypt alerts.",
				Optional:    true,
			},
			"certif_public": schema.StringAttribute{
				Description: "PEM-encoded S/MIME certificate used to encrypt alerts.",
				Optional:    true,
			},
			"expiration": schema.StringAttribute{
				Description: "Account expiry timestamp (RFC3339 or MISP date format).",
				Optional:    true,
			},
		},
	}
}

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateUser(ctx, userFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Creating MISP user failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, userToModel(created))...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetUser(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Reading MISP user failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, userToModel(got))...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateUser(ctx, state.ID.ValueString(), userFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Updating MISP user failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, userToModel(updated))...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteUser(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Deleting MISP user failed", err.Error())
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func userFromModel(m userResourceModel) client.User {
	return client.User{
		Email:         m.Email.ValueString(),
		OrgID:         m.OrgID.ValueString(),
		RoleID:        m.RoleID.ValueString(),
		Autoalert:     m.Autoalert.ValueBool(),
		Contactalert:  m.Contactalert.ValueBool(),
		Disabled:      m.Disabled.ValueBool(),
		Termsaccepted: m.Termsaccepted.ValueBool(),
		ChangePw:      m.ChangePw.ValueBool(),
		GPGKey:        m.GPGKey.ValueString(),
		CertifPublic:  m.CertifPublic.ValueString(),
		Expiration:    m.Expiration.ValueString(),
	}
}

func userToModel(u *client.User) userResourceModel {
	m := userResourceModel{
		ID:            types.StringValue(u.ID),
		Email:         types.StringValue(u.Email),
		OrgID:         types.StringValue(u.OrgID),
		RoleID:        types.StringValue(u.RoleID),
		Autoalert:     types.BoolValue(u.Autoalert),
		Contactalert:  types.BoolValue(u.Contactalert),
		Disabled:      types.BoolValue(u.Disabled),
		Termsaccepted: types.BoolValue(u.Termsaccepted),
		ChangePw:      types.BoolValue(u.ChangePw),
	}
	m.GPGKey = stringOrNull(u.GPGKey)
	m.CertifPublic = stringOrNull(u.CertifPublic)
	m.Expiration = stringOrNull(u.Expiration)
	return m
}
