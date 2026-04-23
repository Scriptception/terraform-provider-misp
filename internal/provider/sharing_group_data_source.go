package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/Scriptception/terraform-provider-misp/internal/client"
)

var (
	_ datasource.DataSource              = (*sharingGroupDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*sharingGroupDataSource)(nil)
)

// NewSharingGroupDataSource constructs a misp_sharing_group data source.
func NewSharingGroupDataSource() datasource.DataSource {
	return &sharingGroupDataSource{}
}

type sharingGroupDataSource struct {
	client *client.Client
}

func (d *sharingGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sharing_group"
}

func (d *sharingGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up a MISP sharing group by id.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Required: true, Description: "Numeric sharing group id."},
			"uuid":          schema.StringAttribute{Computed: true},
			"name":          schema.StringAttribute{Computed: true},
			"description":   schema.StringAttribute{Computed: true},
			"releasability": schema.StringAttribute{Computed: true},
			"org_id":        schema.StringAttribute{Computed: true},
			"active":        schema.BoolAttribute{Computed: true},
			"local":         schema.BoolAttribute{Computed: true},
			"roaming":       schema.BoolAttribute{Computed: true},
		},
	}
}

func (d *sharingGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sharingGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data sharingGroupResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := d.client.GetSharingGroup(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Reading MISP sharing group failed", err.Error())
		return
	}
	out := sgToModel(got)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}
