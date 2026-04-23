package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/Scriptception/terraform-provider-misp/internal/client"
)

var (
	_ datasource.DataSource              = (*warninglistDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*warninglistDataSource)(nil)
)

// NewWarninglistDataSource constructs a misp_warninglist data source. Looks up
// a warninglist by name — the stable, human-readable identifier.
func NewWarninglistDataSource() datasource.DataSource {
	return &warninglistDataSource{}
}

type warninglistDataSource struct {
	client *client.Client
}

func (d *warninglistDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_warninglist"
}

func (d *warninglistDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up a MISP warninglist by name.",
		Attributes: map[string]schema.Attribute{
			"name":                    schema.StringAttribute{Required: true},
			"id":                      schema.StringAttribute{Computed: true},
			"enabled":                 schema.BoolAttribute{Computed: true},
			"description":             schema.StringAttribute{Computed: true},
			"type":                    schema.StringAttribute{Computed: true},
			"version":                 schema.StringAttribute{Computed: true},
			"category":                schema.StringAttribute{Computed: true},
			"valid_attributes":        schema.StringAttribute{Computed: true},
			"warninglist_entry_count": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *warninglistDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *warninglistDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data warninglistResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := d.client.FindWarninglistByName(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Reading MISP warninglist failed", err.Error())
		return
	}
	out := warninglistToModel(got)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}
