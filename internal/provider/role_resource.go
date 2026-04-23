package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Scriptception/terraform-provider-misp/internal/client"
)

var (
	_ resource.Resource                = (*roleResource)(nil)
	_ resource.ResourceWithConfigure   = (*roleResource)(nil)
	_ resource.ResourceWithImportState = (*roleResource)(nil)
)

// NewRoleResource constructs a misp_role resource.
func NewRoleResource() resource.Resource {
	return &roleResource{}
}

type roleResource struct {
	client *client.Client
}

type roleResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Permission            types.String `tfsdk:"permission"`
	// Derived from permission (Computed-only)
	PermAdd               types.Bool   `tfsdk:"perm_add"`
	PermModify            types.Bool   `tfsdk:"perm_modify"`
	PermModifyOrg         types.Bool   `tfsdk:"perm_modify_org"`
	PermPublish           types.Bool   `tfsdk:"perm_publish"`
	PermDelegate          types.Bool   `tfsdk:"perm_delegate"`
	// Independent booleans
	PermSighting          types.Bool   `tfsdk:"perm_sighting"`
	PermTagger            types.Bool   `tfsdk:"perm_tagger"`
	PermTemplate          types.Bool   `tfsdk:"perm_template"`
	PermSharingGroup      types.Bool   `tfsdk:"perm_sharing_group"`
	PermTagEditor         types.Bool   `tfsdk:"perm_tag_editor"`
	PermObjectTemplate    types.Bool   `tfsdk:"perm_object_template"`
	PermSync              types.Bool   `tfsdk:"perm_sync"`
	PermAdmin             types.Bool   `tfsdk:"perm_admin"`
	PermAuth              types.Bool   `tfsdk:"perm_auth"`
	PermSiteAdmin         types.Bool   `tfsdk:"perm_site_admin"`
	PermRegexpAccess      types.Bool   `tfsdk:"perm_regexp_access"`
	PermAudit             types.Bool   `tfsdk:"perm_audit"`
	PermPublishZmq        types.Bool   `tfsdk:"perm_publish_zmq"`
	PermPublishKafka      types.Bool   `tfsdk:"perm_publish_kafka"`
	PermDecaying          types.Bool   `tfsdk:"perm_decaying"`
	PermGalaxyEditor      types.Bool   `tfsdk:"perm_galaxy_editor"`
	DefaultRole           types.Bool   `tfsdk:"default_role"`
	RestrictedToSiteAdmin types.Bool   `tfsdk:"restricted_to_site_admin"`
	EnforceRateLimit      types.Bool   `tfsdk:"enforce_rate_limit"`
	MemoryLimit           types.String `tfsdk:"memory_limit"`
	MaxExecutionTime      types.String `tfsdk:"max_execution_time"`
	RateLimitCount        types.String `tfsdk:"rate_limit_count"`
}

func (r *roleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *roleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A MISP role. Controls what actions users assigned to this role may perform. " +
			"The `permission` field (\"0\"–\"3\") governs the five derived publish/modify flags; " +
			"all other `perm_*` flags are independent and persisted as set.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Numeric MISP identifier.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Role name. Must be unique within the MISP instance.",
				Required:    true,
			},
			"permission": schema.StringAttribute{
				Description: "Base permission level: \"0\"=read-only, \"1\"=add, \"2\"=add+modify, \"3\"=add+modify+publish+delegate. " +
					"Controls the five derived perm_add/perm_modify/perm_modify_org/perm_publish/perm_delegate flags.",
				Required: true,
				Validators: []validator.String{
					permissionEnumValidator{},
				},
			},
			// ---- Derived (Computed-only) ----
			"perm_add": schema.BoolAttribute{
				Description: "Computed from permission>=1. Do not set directly.",
				Computed:    true,
			},
			"perm_modify": schema.BoolAttribute{
				Description: "Computed from permission>=2. Do not set directly.",
				Computed:    true,
			},
			"perm_modify_org": schema.BoolAttribute{
				Description: "Computed from permission>=2. Do not set directly.",
				Computed:    true,
			},
			"perm_publish": schema.BoolAttribute{
				Description: "Computed from permission>=3. Do not set directly.",
				Computed:    true,
			},
			"perm_delegate": schema.BoolAttribute{
				Description: "Computed from permission>=3. Do not set directly.",
				Computed:    true,
			},
			// ---- Independent booleans ----
			"perm_sighting": schema.BoolAttribute{
				Description: "Allow sighting.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"perm_tagger": schema.BoolAttribute{
				Description: "Allow tagging events/attributes.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"perm_template": schema.BoolAttribute{
				Description: "Allow managing event templates.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"perm_sharing_group": schema.BoolAttribute{
				Description: "Allow managing sharing groups.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"perm_tag_editor": schema.BoolAttribute{
				Description: "Allow creating and editing tags.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"perm_object_template": schema.BoolAttribute{
				Description: "Allow managing object templates.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"perm_sync": schema.BoolAttribute{
				Description: "Allow synchronisation actions.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"perm_admin": schema.BoolAttribute{
				Description: "Allow organisation-admin actions.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"perm_auth": schema.BoolAttribute{
				Description: "Allow use of authentication keys.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"perm_site_admin": schema.BoolAttribute{
				Description: "Grant site-admin privileges.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"perm_regexp_access": schema.BoolAttribute{
				Description: "Allow use of regular-expression clean-up tools.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"perm_audit": schema.BoolAttribute{
				Description: "Allow access to the audit log.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"perm_publish_zmq": schema.BoolAttribute{
				Description: "Allow publishing to the ZMQ feed.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"perm_publish_kafka": schema.BoolAttribute{
				Description: "Allow publishing to the Kafka feed.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"perm_decaying": schema.BoolAttribute{
				Description: "Allow managing decaying models.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"perm_galaxy_editor": schema.BoolAttribute{
				Description: "Allow editing galaxies and clusters.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"default_role": schema.BoolAttribute{
				Description: "Make this the default role for new users.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"restricted_to_site_admin": schema.BoolAttribute{
				Description: "Restrict membership of this role to site admins only.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"enforce_rate_limit": schema.BoolAttribute{
				Description: "Enable rate limiting for users with this role.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			// ---- String options ----
			"memory_limit": schema.StringAttribute{
				Description: "PHP memory limit for users in this role (e.g. \"512M\"). Empty string means no override.",
				Optional:    true,
				Computed:    true,
			},
			"max_execution_time": schema.StringAttribute{
				Description: "PHP max_execution_time for users in this role (seconds as string). Empty string means no override.",
				Optional:    true,
				Computed:    true,
			},
			"rate_limit_count": schema.StringAttribute{
				Description: "API requests allowed per rate-limit window (numeric string).",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *roleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan roleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateRole(ctx, roleFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Creating MISP role failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, roleToModel(created))...)
}

func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetRole(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Reading MISP role failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, roleToModel(got))...)
}

func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan roleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state roleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateRole(ctx, state.ID.ValueString(), roleFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Updating MISP role failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, roleToModel(updated))...)
}

func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state roleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteRole(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Deleting MISP role failed", err.Error())
	}
}

func (r *roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// roleFromModel converts a Terraform model into a client Role struct.
// The five derived perm_* fields are intentionally omitted: MISP derives them
// from permission and any value we send would be ignored.
func roleFromModel(m roleResourceModel) client.Role {
	fb := func(b types.Bool) client.FlexBool { return client.FlexBool(b.ValueBool()) }
	return client.Role{
		Name:                  m.Name.ValueString(),
		Permission:            m.Permission.ValueString(),
		PermSighting:          fb(m.PermSighting),
		PermTagger:            fb(m.PermTagger),
		PermTemplate:          fb(m.PermTemplate),
		PermSharingGroup:      fb(m.PermSharingGroup),
		PermTagEditor:         fb(m.PermTagEditor),
		PermObjectTemplate:    fb(m.PermObjectTemplate),
		PermSync:              fb(m.PermSync),
		PermAdmin:             fb(m.PermAdmin),
		PermAuth:              fb(m.PermAuth),
		PermSiteAdmin:         fb(m.PermSiteAdmin),
		PermRegexpAccess:      fb(m.PermRegexpAccess),
		PermAudit:             fb(m.PermAudit),
		PermPublishZmq:        fb(m.PermPublishZmq),
		PermPublishKafka:      fb(m.PermPublishKafka),
		PermDecaying:          fb(m.PermDecaying),
		PermGalaxyEditor:      fb(m.PermGalaxyEditor),
		DefaultRole:           fb(m.DefaultRole),
		RestrictedToSiteAdmin: fb(m.RestrictedToSiteAdmin),
		EnforceRateLimit:      fb(m.EnforceRateLimit),
		MemoryLimit:           client.FlexString(m.MemoryLimit.ValueString()),
		MaxExecutionTime:      client.FlexString(m.MaxExecutionTime.ValueString()),
		RateLimitCount:        client.FlexString(m.RateLimitCount.ValueString()),
	}
}

// roleToModel converts a client Role struct into the Terraform state model.
func roleToModel(r *client.Role) roleResourceModel {
	tb := func(b client.FlexBool) types.Bool { return types.BoolValue(b.Bool()) }
	return roleResourceModel{
		ID:                    types.StringValue(r.ID),
		Name:                  types.StringValue(r.Name),
		Permission:            types.StringValue(r.Permission),
		PermAdd:               tb(r.PermAdd),
		PermModify:            tb(r.PermModify),
		PermModifyOrg:         tb(r.PermModifyOrg),
		PermPublish:           tb(r.PermPublish),
		PermDelegate:          tb(r.PermDelegate),
		PermSighting:          tb(r.PermSighting),
		PermTagger:            tb(r.PermTagger),
		PermTemplate:          tb(r.PermTemplate),
		PermSharingGroup:      tb(r.PermSharingGroup),
		PermTagEditor:         tb(r.PermTagEditor),
		PermObjectTemplate:    tb(r.PermObjectTemplate),
		PermSync:              tb(r.PermSync),
		PermAdmin:             tb(r.PermAdmin),
		PermAuth:              tb(r.PermAuth),
		PermSiteAdmin:         tb(r.PermSiteAdmin),
		PermRegexpAccess:      tb(r.PermRegexpAccess),
		PermAudit:             tb(r.PermAudit),
		PermPublishZmq:        tb(r.PermPublishZmq),
		PermPublishKafka:      tb(r.PermPublishKafka),
		PermDecaying:          tb(r.PermDecaying),
		PermGalaxyEditor:      tb(r.PermGalaxyEditor),
		DefaultRole:           tb(r.DefaultRole),
		RestrictedToSiteAdmin: tb(r.RestrictedToSiteAdmin),
		EnforceRateLimit:      tb(r.EnforceRateLimit),
		MemoryLimit:           stringOrNull(r.MemoryLimit.String()),
		MaxExecutionTime:      stringOrNull(r.MaxExecutionTime.String()),
		RateLimitCount:        stringOrNull(r.RateLimitCount.String()),
	}
}

// permissionEnumValidator validates that the permission field is one of "0","1","2","3".
type permissionEnumValidator struct{}

func (v permissionEnumValidator) Description(_ context.Context) string {
	return `must be one of "0", "1", "2", or "3"`
}

func (v permissionEnumValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v permissionEnumValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueString()
	for _, allowed := range []string{"0", "1", "2", "3"} {
		if val == allowed {
			return
		}
	}
	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid permission value",
		fmt.Sprintf("permission must be one of \"0\", \"1\", \"2\", or \"3\"; got %q", val),
	)
}
