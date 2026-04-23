package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Scriptception/terraform-provider-misp/internal/client"
)

var (
	_ datasource.DataSource              = (*galaxyClusterDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*galaxyClusterDataSource)(nil)
)

// NewGalaxyClusterDataSource constructs a misp_galaxy_cluster data source.
func NewGalaxyClusterDataSource() datasource.DataSource {
	return &galaxyClusterDataSource{}
}

type galaxyClusterDataSource struct {
	client *client.Client
}

func (d *galaxyClusterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_galaxy_cluster"
}

func (d *galaxyClusterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up a MISP galaxy cluster by id. Works for both custom and bundled clusters (read-only).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Numeric galaxy cluster id.",
			},
			"galaxy_id":        schema.StringAttribute{Computed: true},
			"uuid":             schema.StringAttribute{Computed: true},
			"collection_uuid":  schema.StringAttribute{Computed: true},
			"type":             schema.StringAttribute{Computed: true},
			"value":            schema.StringAttribute{Computed: true},
			"tag_name":         schema.StringAttribute{Computed: true},
			"description":      schema.StringAttribute{Computed: true},
			"source":           schema.StringAttribute{Computed: true},
			"version":          schema.StringAttribute{Computed: true},
			"distribution":     schema.StringAttribute{Computed: true},
			"sharing_group_id": schema.StringAttribute{Computed: true},
			"org_id":           schema.StringAttribute{Computed: true},
			"orgc_id":          schema.StringAttribute{Computed: true},
			"locked":           schema.BoolAttribute{Computed: true},
			"extends_uuid":     schema.StringAttribute{Computed: true},
			"extends_version":  schema.StringAttribute{Computed: true},
			"published":        schema.BoolAttribute{Computed: true},
			"authors": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"elements": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Key-value metadata for this cluster.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key":   schema.StringAttribute{Computed: true},
						"value": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *galaxyClusterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *galaxyClusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data galaxyClusterResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	got, err := d.client.GetGalaxyCluster(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Reading MISP galaxy cluster failed", err.Error())
		return
	}

	out, diags := gcToModel(ctx, got)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}
