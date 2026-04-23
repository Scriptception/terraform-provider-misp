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
	_ datasource.DataSource              = (*organisationDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*organisationDataSource)(nil)
)

// NewOrganisationDataSource constructs a misp_organisation data source.
func NewOrganisationDataSource() datasource.DataSource {
	return &organisationDataSource{}
}

type organisationDataSource struct {
	client *client.Client
}

func (d *organisationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organisation"
}

func (d *organisationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up a MISP organisation by id or uuid.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Numeric id or uuid to look up.",
				Required:    true,
			},
			"uuid":        schema.StringAttribute{Computed: true},
			"name":        schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
			"type":        schema.StringAttribute{Computed: true},
			"nationality": schema.StringAttribute{Computed: true},
			"sector":      schema.StringAttribute{Computed: true},
			"contacts":    schema.StringAttribute{Computed: true},
			"local": schema.BoolAttribute{Computed: true},
			"restricted_to_domain": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *organisationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *organisationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data organisationResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	got, err := d.client.GetOrganisation(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Reading MISP organisation failed", err.Error())
		return
	}
	out := orgToModel(got)
	out.ID = types.StringValue(got.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}
