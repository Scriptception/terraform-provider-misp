package provider

// # Future work (deferred for v0.2)
//
// The following Server fields are intentionally omitted from this resource:
//   - push_rules / pull_rules: stringified JSON blobs that require special
//     diff-suppression logic to prevent spurious plan changes.
//   - cert_file / client_cert_file: base64-encoded PEM that MISP marks
//     Sensitive and omits from GET responses (write-only semantics).
//   - lastpulledid / lastpushedid: operational counters managed by MISP,
//     not meaningful to declare in Terraform config.

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Scriptception/terraform-provider-misp/internal/client"
)

var (
	_ resource.Resource                = (*serverResource)(nil)
	_ resource.ResourceWithConfigure   = (*serverResource)(nil)
	_ resource.ResourceWithImportState = (*serverResource)(nil)
)

// NewServerResource constructs a misp_server resource.
func NewServerResource() resource.Resource {
	return &serverResource{}
}

type serverResource struct {
	client *client.Client
}

type serverResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	URL                 types.String `tfsdk:"url"`
	Authkey             types.String `tfsdk:"authkey"`
	RemoteOrgID         types.String `tfsdk:"remote_org_id"`
	Push                types.Bool   `tfsdk:"push"`
	Pull                types.Bool   `tfsdk:"pull"`
	PushSightings       types.Bool   `tfsdk:"push_sightings"`
	PushGalaxyClusters  types.Bool   `tfsdk:"push_galaxy_clusters"`
	PullGalaxyClusters  types.Bool   `tfsdk:"pull_galaxy_clusters"`
	SelfSigned          types.Bool   `tfsdk:"self_signed"`
	SkipProxy           types.Bool   `tfsdk:"skip_proxy"`
	CachingEnabled      types.Bool   `tfsdk:"caching_enabled"`
	UnpublishEvent      types.Bool   `tfsdk:"unpublish_event"`
	PublishWithoutEmail types.Bool   `tfsdk:"publish_without_email"`
	Internal            types.Bool   `tfsdk:"internal"`
	Organization        types.String `tfsdk:"organization"`
}

func (r *serverResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *serverResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A MISP sync server (peer). Represents a remote MISP instance that this instance can push to or pull from.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Numeric MISP identifier.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Human-readable label for the remote server.",
				Required:    true,
			},
			"url": schema.StringAttribute{
				Description: "Base URL of the remote MISP instance (e.g. `https://misp-peer.example.com`).",
				Required:    true,
			},
			"authkey": schema.StringAttribute{
				Description: "API key used to authenticate against the remote server. Sensitive; not returned by MISP on read.",
				Required:    true,
				Sensitive:   true,
			},
			"remote_org_id": schema.StringAttribute{
				Description: "Numeric id of the organisation on *this* instance that represents the remote server's owner.",
				Required:    true,
			},
			"push": schema.BoolAttribute{
				Description: "Enable pushing events to this server.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"pull": schema.BoolAttribute{
				Description: "Enable pulling events from this server.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"push_sightings": schema.BoolAttribute{
				Description: "Push sightings to this server.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"push_galaxy_clusters": schema.BoolAttribute{
				Description: "Push galaxy clusters to this server.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"pull_galaxy_clusters": schema.BoolAttribute{
				Description: "Pull galaxy clusters from this server.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"self_signed": schema.BoolAttribute{
				Description: "Accept self-signed TLS certificates from the remote server.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"skip_proxy": schema.BoolAttribute{
				Description: "Bypass the configured proxy when connecting to this server.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"caching_enabled": schema.BoolAttribute{
				Description: "Cache data from this server for faster lookups.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"unpublish_event": schema.BoolAttribute{
				Description: "Unpublish events after pulling them from this server.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"publish_without_email": schema.BoolAttribute{
				Description: "Publish pulled events without sending notification emails.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"internal": schema.BoolAttribute{
				Description: "Mark this server as an internal link (affects distribution handling).",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"organization": schema.StringAttribute{
				Description: "Name of the remote organisation as learned by MISP after the first sync. Computed.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *serverResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *serverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serverResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateServer(ctx, serverFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Creating MISP server failed", err.Error())
		return
	}
	// Preserve the authkey from the plan: MISP does not echo it back on read.
	state := serverToModel(created)
	state.Authkey = plan.Authkey
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *serverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serverResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetServer(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Reading MISP server failed", err.Error())
		return
	}
	fresh := serverToModel(got)
	// authkey is not returned by MISP on read; carry the last known value
	// forward so Terraform doesn't see a spurious diff.
	fresh.Authkey = state.Authkey
	resp.Diagnostics.Append(resp.State.Set(ctx, fresh)...)
}

func (r *serverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serverResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serverResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateServer(ctx, state.ID.ValueString(), serverFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Updating MISP server failed", err.Error())
		return
	}
	fresh := serverToModel(updated)
	// Preserve the authkey from the plan; MISP does not echo it back.
	fresh.Authkey = plan.Authkey
	resp.Diagnostics.Append(resp.State.Set(ctx, fresh)...)
}

func (r *serverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serverResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteServer(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Deleting MISP server failed", err.Error())
	}
}

func (r *serverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// serverFromModel converts a Terraform model into the client struct sent to
// MISP. The id field is intentionally omitted so it does not appear in
// create/update request bodies.
func serverFromModel(m serverResourceModel) client.Server {
	return client.Server{
		Name:                m.Name.ValueString(),
		URL:                 m.URL.ValueString(),
		Authkey:             m.Authkey.ValueString(),
		RemoteOrgID:         m.RemoteOrgID.ValueString(),
		Push:                m.Push.ValueBool(),
		Pull:                m.Pull.ValueBool(),
		PushSightings:       m.PushSightings.ValueBool(),
		PushGalaxyClusters:  m.PushGalaxyClusters.ValueBool(),
		PullGalaxyClusters:  m.PullGalaxyClusters.ValueBool(),
		SelfSigned:          m.SelfSigned.ValueBool(),
		SkipProxy:           m.SkipProxy.ValueBool(),
		CachingEnabled:      m.CachingEnabled.ValueBool(),
		UnpublishEvent:      m.UnpublishEvent.ValueBool(),
		PublishWithoutEmail: m.PublishWithoutEmail.ValueBool(),
		Internal:            m.Internal.ValueBool(),
	}
}

// serverToModel converts a client struct received from MISP into the
// Terraform model. The authkey is NOT set here because MISP does not return
// it on read; callers are responsible for preserving it from prior state.
func serverToModel(s *client.Server) serverResourceModel {
	return serverResourceModel{
		ID:                  types.StringValue(s.ID),
		Name:                types.StringValue(s.Name),
		URL:                 types.StringValue(s.URL),
		Authkey:             types.StringValue(""), // not returned by MISP; caller sets from plan/state
		RemoteOrgID:         types.StringValue(s.RemoteOrgID),
		Push:                types.BoolValue(s.Push),
		Pull:                types.BoolValue(s.Pull),
		PushSightings:       types.BoolValue(s.PushSightings),
		PushGalaxyClusters:  types.BoolValue(s.PushGalaxyClusters),
		PullGalaxyClusters:  types.BoolValue(s.PullGalaxyClusters),
		SelfSigned:          types.BoolValue(s.SelfSigned),
		SkipProxy:           types.BoolValue(s.SkipProxy),
		CachingEnabled:      types.BoolValue(s.CachingEnabled),
		UnpublishEvent:      types.BoolValue(s.UnpublishEvent),
		PublishWithoutEmail: types.BoolValue(s.PublishWithoutEmail),
		Internal:            types.BoolValue(s.Internal),
		Organization:        types.StringValue(s.Organization),
	}
}
