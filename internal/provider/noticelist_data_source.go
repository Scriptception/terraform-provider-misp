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
	_ datasource.DataSource              = (*noticelistDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*noticelistDataSource)(nil)
)

// NewNoticelistDataSource constructs a misp_noticelist data source. Looks up
// a noticelist by name — the stable, human-readable identifier.
func NewNoticelistDataSource() datasource.DataSource {
	return &noticelistDataSource{}
}

type noticelistDataSource struct {
	client *client.Client
}

func (d *noticelistDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_noticelist"
}

func (d *noticelistDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up a MISP noticelist by name.",
		Attributes: map[string]schema.Attribute{
			"name":          schema.StringAttribute{Required: true},
			"id":            schema.StringAttribute{Computed: true},
			"enabled":       schema.BoolAttribute{Computed: true},
			"expanded_name": schema.StringAttribute{Computed: true},
			"version":       schema.StringAttribute{Computed: true},
			"ref": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"geographical_area": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *noticelistDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *noticelistDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data noticelistResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := d.client.FindNoticelistByName(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Reading MISP noticelist failed", err.Error())
		return
	}
	out, diags := noticelistToModel(ctx, got)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}
