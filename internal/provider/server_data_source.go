package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/Scriptception/terraform-provider-misp/internal/client"
)

var (
	_ datasource.DataSource              = (*serverDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*serverDataSource)(nil)
)

// NewServerDataSource constructs a misp_server data source.
func NewServerDataSource() datasource.DataSource {
	return &serverDataSource{}
}

type serverDataSource struct {
	client *client.Client
}

func (d *serverDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (d *serverDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up a MISP sync server by its numeric id.",
		Attributes: map[string]schema.Attribute{
			"id":                   schema.StringAttribute{Required: true, Description: "Numeric server id."},
			"name":                 schema.StringAttribute{Computed: true},
			"url":                  schema.StringAttribute{Computed: true},
			"authkey":              schema.StringAttribute{Computed: true, Sensitive: true},
			"remote_org_id":        schema.StringAttribute{Computed: true},
			"push":                 schema.BoolAttribute{Computed: true},
			"pull":                 schema.BoolAttribute{Computed: true},
			"push_sightings":       schema.BoolAttribute{Computed: true},
			"push_galaxy_clusters": schema.BoolAttribute{Computed: true},
			"pull_galaxy_clusters": schema.BoolAttribute{Computed: true},
			"self_signed":          schema.BoolAttribute{Computed: true},
			"skip_proxy":           schema.BoolAttribute{Computed: true},
			"caching_enabled":      schema.BoolAttribute{Computed: true},
			"unpublish_event":      schema.BoolAttribute{Computed: true},
			"publish_without_email": schema.BoolAttribute{Computed: true},
			"internal":             schema.BoolAttribute{Computed: true},
			"organization":         schema.StringAttribute{Computed: true},
		},
	}
}

func (d *serverDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *serverDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data serverResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := d.client.GetServer(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Reading MISP server failed", err.Error())
		return
	}
	out := serverToModel(got)
	// authkey is not returned by MISP on read; leave as empty string.
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}
