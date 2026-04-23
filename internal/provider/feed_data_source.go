package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/Scriptception/terraform-provider-misp/internal/client"
)

var (
	_ datasource.DataSource              = (*feedDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*feedDataSource)(nil)
)

// NewFeedDataSource constructs a misp_feed data source.
func NewFeedDataSource() datasource.DataSource {
	return &feedDataSource{}
}

type feedDataSource struct {
	client *client.Client
}

func (d *feedDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_feed"
}

func (d *feedDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up a MISP feed by id.",
		Attributes: map[string]schema.Attribute{
			"id":               schema.StringAttribute{Required: true, Description: "Numeric feed id."},
			"name":             schema.StringAttribute{Computed: true},
			"provider_name":    schema.StringAttribute{Computed: true},
			"url":              schema.StringAttribute{Computed: true},
			"source_format":    schema.StringAttribute{Computed: true},
			"enabled":          schema.BoolAttribute{Computed: true},
			"distribution":     schema.StringAttribute{Computed: true},
			"sharing_group_id": schema.StringAttribute{Computed: true},
			"tag_id":           schema.StringAttribute{Computed: true},
			"orgc_id":          schema.StringAttribute{Computed: true},
			"fixed_event":      schema.BoolAttribute{Computed: true},
			"delta_merge":      schema.BoolAttribute{Computed: true},
			"publish":          schema.BoolAttribute{Computed: true},
			"override_ids":     schema.BoolAttribute{Computed: true},
			"caching_enabled":  schema.BoolAttribute{Computed: true},
			"force_to_ids":     schema.BoolAttribute{Computed: true},
			"lookup_visible":   schema.BoolAttribute{Computed: true},
			"input_source":     schema.StringAttribute{Computed: true},
			"rules":            schema.StringAttribute{Computed: true},
		},
	}
}

func (d *feedDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("expected *client.Client, got %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *feedDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data feedResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := d.client.GetFeed(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Reading MISP feed failed", err.Error())
		return
	}
	out := feedToModel(got)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}
