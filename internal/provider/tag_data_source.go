package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/Scriptception/terraform-provider-misp/internal/client"
)

var (
	_ datasource.DataSource              = (*tagDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*tagDataSource)(nil)
)

// NewTagDataSource constructs a misp_tag data source.
func NewTagDataSource() datasource.DataSource {
	return &tagDataSource{}
}

type tagDataSource struct {
	client *client.Client
}

func (d *tagDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

func (d *tagDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up a MISP tag by id.",
		Attributes: map[string]schema.Attribute{
			"id":              schema.StringAttribute{Required: true, Description: "Numeric tag id."},
			"name":            schema.StringAttribute{Computed: true},
			"colour":          schema.StringAttribute{Computed: true},
			"exportable":      schema.BoolAttribute{Computed: true},
			"hide_tag":        schema.BoolAttribute{Computed: true},
			"numerical_value": schema.Int64Attribute{Computed: true},
			"org_id":          schema.StringAttribute{Computed: true},
			"user_id":         schema.StringAttribute{Computed: true},
			"local_only":      schema.BoolAttribute{Computed: true},
		},
	}
}

func (d *tagDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *tagDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data tagResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := d.client.GetTag(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Reading MISP tag failed", err.Error())
		return
	}
	out := tagToModel(got)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}
