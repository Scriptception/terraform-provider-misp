package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/Scriptception/terraform-provider-misp/internal/client"
)

var (
	_ datasource.DataSource              = (*taxonomyDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*taxonomyDataSource)(nil)
)

// NewTaxonomyDataSource constructs a misp_taxonomy data source. Looks up
// a taxonomy by namespace — the stable, human-readable identifier.
func NewTaxonomyDataSource() datasource.DataSource {
	return &taxonomyDataSource{}
}

type taxonomyDataSource struct {
	client *client.Client
}

func (d *taxonomyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_taxonomy"
}

func (d *taxonomyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up a MISP taxonomy by namespace.",
		Attributes: map[string]schema.Attribute{
			"namespace":   schema.StringAttribute{Required: true},
			"id":          schema.StringAttribute{Computed: true},
			"enabled":     schema.BoolAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
			"version":     schema.StringAttribute{Computed: true},
			"exclusive":   schema.BoolAttribute{Computed: true},
			"required":    schema.BoolAttribute{Computed: true},
		},
	}
}

func (d *taxonomyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *taxonomyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data taxonomyResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := d.client.FindTaxonomyByNamespace(ctx, data.Namespace.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Reading MISP taxonomy failed", err.Error())
		return
	}
	out := taxonomyToModel(got)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}
