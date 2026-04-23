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
	_ resource.Resource                = (*feedResource)(nil)
	_ resource.ResourceWithConfigure   = (*feedResource)(nil)
	_ resource.ResourceWithImportState = (*feedResource)(nil)
)

// NewFeedResource constructs a misp_feed resource.
func NewFeedResource() resource.Resource {
	return &feedResource{}
}

type feedResource struct {
	client *client.Client
}

type feedResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Provider       types.String `tfsdk:"provider_name"`
	URL            types.String `tfsdk:"url"`
	SourceFormat   types.String `tfsdk:"source_format"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	Distribution   types.String `tfsdk:"distribution"`
	SharingGroupID types.String `tfsdk:"sharing_group_id"`
	TagID          types.String `tfsdk:"tag_id"`
	OrgcID         types.String `tfsdk:"orgc_id"`
	FixedEvent     types.Bool   `tfsdk:"fixed_event"`
	DeltaMerge     types.Bool   `tfsdk:"delta_merge"`
	Publish        types.Bool   `tfsdk:"publish"`
	OverrideIDs    types.Bool   `tfsdk:"override_ids"`
	CachingEnabled types.Bool   `tfsdk:"caching_enabled"`
	ForceToIDs     types.Bool   `tfsdk:"force_to_ids"`
	LookupVisible  types.Bool   `tfsdk:"lookup_visible"`
	InputSource    types.String `tfsdk:"input_source"`
	Rules          types.String `tfsdk:"rules"`
}

func (r *feedResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_feed"
}

func (r *feedResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A MISP feed. Feeds allow MISP to pull threat intelligence from external sources on a schedule.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Numeric MISP identifier.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Feed name.",
				Required:    true,
			},
			"provider_name": schema.StringAttribute{
				Description: "Human-readable name of the feed's upstream provider (e.g. `abuse.ch`). Named `provider_name` because `provider` is a reserved word in Terraform — the underlying MISP field is still `provider`.",
				Required:    true,
			},
			"url": schema.StringAttribute{
				Description: "URL of the feed source.",
				Required:    true,
			},
			"source_format": schema.StringAttribute{
				Description: "Feed format. Common values: `misp`, `freetext`, `csv`. Defaults to `misp`.",
				Optional:    true,
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the feed is active and will be fetched on schedule.",
				Optional:    true,
				Computed:    true,
			},
			"distribution": schema.StringAttribute{
				Description: "Distribution level (`0`=org only, `1`=community, `2`=connected, `3`=all).",
				Optional:    true,
				Computed:    true,
			},
			"sharing_group_id": schema.StringAttribute{
				Description: "Sharing group id (used when distribution is `4`).",
				Optional:    true,
				Computed:    true,
			},
			"tag_id": schema.StringAttribute{
				Description: "Tag id to apply to events pulled from this feed.",
				Optional:    true,
				Computed:    true,
			},
			"orgc_id": schema.StringAttribute{
				Description: "Creator organisation id to attribute to imported events.",
				Optional:    true,
				Computed:    true,
			},
			"fixed_event": schema.BoolAttribute{
				Description: "When true, all feed data is merged into a single fixed event instead of creating new events.",
				Optional:    true,
				Computed:    true,
			},
			"delta_merge": schema.BoolAttribute{
				Description: "When true, attributes removed from the feed are also removed from MISP.",
				Optional:    true,
				Computed:    true,
			},
			"publish": schema.BoolAttribute{
				Description: "Publish events imported from this feed immediately.",
				Optional:    true,
				Computed:    true,
			},
			"override_ids": schema.BoolAttribute{
				Description: "Override the IDS flag on imported attributes.",
				Optional:    true,
				Computed:    true,
			},
			"caching_enabled": schema.BoolAttribute{
				Description: "Enable local caching of feed data.",
				Optional:    true,
				Computed:    true,
			},
			"force_to_ids": schema.BoolAttribute{
				Description: "Force all imported attributes to be marked as IDS signatures.",
				Optional:    true,
				Computed:    true,
			},
			"lookup_visible": schema.BoolAttribute{
				Description: "Make the feed visible in the lookup interface.",
				Optional:    true,
				Computed:    true,
			},
			"input_source": schema.StringAttribute{
				Description: "Source type: `network` (fetched over HTTP/S) or `local` (local file path).",
				Optional:    true,
				Computed:    true,
			},
			"rules": schema.StringAttribute{
				Description: "JSON-encoded filtering rules applied when importing from the feed.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *feedResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *feedResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan feedResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateFeed(ctx, feedFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Creating MISP feed failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, feedToModel(created))...)
}

func (r *feedResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state feedResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetFeed(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Reading MISP feed failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, feedToModel(got))...)
}

func (r *feedResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan feedResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state feedResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateFeed(ctx, state.ID.ValueString(), feedFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Updating MISP feed failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, feedToModel(updated))...)
}

func (r *feedResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state feedResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteFeed(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Deleting MISP feed failed", err.Error())
	}
}

func (r *feedResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func feedFromModel(m feedResourceModel) client.Feed {
	return client.Feed{
		Name:           m.Name.ValueString(),
		Provider:       m.Provider.ValueString(),
		URL:            m.URL.ValueString(),
		SourceFormat:   m.SourceFormat.ValueString(),
		Enabled:        m.Enabled.ValueBool(),
		Distribution:   m.Distribution.ValueString(),
		SharingGroupID: m.SharingGroupID.ValueString(),
		TagID:          m.TagID.ValueString(),
		OrgcID:         m.OrgcID.ValueString(),
		FixedEvent:     m.FixedEvent.ValueBool(),
		DeltaMerge:     m.DeltaMerge.ValueBool(),
		Publish:        m.Publish.ValueBool(),
		OverrideIDs:    m.OverrideIDs.ValueBool(),
		CachingEnabled: m.CachingEnabled.ValueBool(),
		ForceToIDs:     m.ForceToIDs.ValueBool(),
		LookupVisible:  m.LookupVisible.ValueBool(),
		InputSource:    m.InputSource.ValueString(),
		Rules:          m.Rules.ValueString(),
	}
}

func feedToModel(f *client.Feed) feedResourceModel {
	m := feedResourceModel{
		ID:             types.StringValue(f.ID),
		Name:           types.StringValue(f.Name),
		Provider:       types.StringValue(f.Provider),
		URL:            types.StringValue(f.URL),
		SourceFormat:   types.StringValue(f.SourceFormat),
		Enabled:        types.BoolValue(f.Enabled),
		Distribution:   types.StringValue(f.Distribution),
		SharingGroupID: types.StringValue(f.SharingGroupID),
		TagID:          types.StringValue(f.TagID),
		OrgcID:         types.StringValue(f.OrgcID),
		FixedEvent:     types.BoolValue(f.FixedEvent),
		DeltaMerge:     types.BoolValue(f.DeltaMerge),
		Publish:        types.BoolValue(f.Publish),
		OverrideIDs:    types.BoolValue(f.OverrideIDs),
		CachingEnabled: types.BoolValue(f.CachingEnabled),
		ForceToIDs:     types.BoolValue(f.ForceToIDs),
		LookupVisible:  types.BoolValue(f.LookupVisible),
		InputSource:    types.StringValue(f.InputSource),
	}
	m.Rules = stringOrNull(f.Rules)
	return m
}
