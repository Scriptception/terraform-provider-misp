package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/Scriptception/terraform-provider-misp/internal/client"
)

var (
	_ datasource.DataSource              = (*settingDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*settingDataSource)(nil)
)

// NewSettingDataSource constructs a misp_setting data source. Looks up a
// MISP server setting by its dotted name (e.g. "MISP.baseurl").
func NewSettingDataSource() datasource.DataSource {
	return &settingDataSource{}
}

type settingDataSource struct {
	client *client.Client
}

func (d *settingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_setting"
}

func (d *settingDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Read the current value of a MISP server setting by its dotted name (e.g. `MISP.baseurl`).",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true},
			"name":        schema.StringAttribute{Required: true},
			"value":       schema.StringAttribute{Computed: true},
			"type":        schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
			"level":       schema.Int64Attribute{Computed: true},
		},
	}
}

func (d *settingDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *settingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data settingResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := d.client.GetSetting(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Reading MISP setting failed", err.Error())
		return
	}
	out := settingToModel(got)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}
