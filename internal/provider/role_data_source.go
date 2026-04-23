package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/Scriptception/terraform-provider-misp/internal/client"
)

var (
	_ datasource.DataSource              = (*roleDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*roleDataSource)(nil)
)

// NewRoleDataSource constructs a misp_role data source.
func NewRoleDataSource() datasource.DataSource {
	return &roleDataSource{}
}

type roleDataSource struct {
	client *client.Client
}

func (d *roleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (d *roleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up a MISP role by id.",
		Attributes: map[string]schema.Attribute{
			"id":                     schema.StringAttribute{Required: true, Description: "Numeric role id."},
			"name":                   schema.StringAttribute{Computed: true},
			"permission":             schema.StringAttribute{Computed: true},
			"perm_add":               schema.BoolAttribute{Computed: true},
			"perm_modify":            schema.BoolAttribute{Computed: true},
			"perm_modify_org":        schema.BoolAttribute{Computed: true},
			"perm_publish":           schema.BoolAttribute{Computed: true},
			"perm_delegate":          schema.BoolAttribute{Computed: true},
			"perm_sighting":          schema.BoolAttribute{Computed: true},
			"perm_tagger":            schema.BoolAttribute{Computed: true},
			"perm_template":          schema.BoolAttribute{Computed: true},
			"perm_sharing_group":     schema.BoolAttribute{Computed: true},
			"perm_tag_editor":        schema.BoolAttribute{Computed: true},
			"perm_object_template":   schema.BoolAttribute{Computed: true},
			"perm_sync":              schema.BoolAttribute{Computed: true},
			"perm_admin":             schema.BoolAttribute{Computed: true},
			"perm_auth":              schema.BoolAttribute{Computed: true},
			"perm_site_admin":        schema.BoolAttribute{Computed: true},
			"perm_regexp_access":     schema.BoolAttribute{Computed: true},
			"perm_audit":             schema.BoolAttribute{Computed: true},
			"perm_publish_zmq":       schema.BoolAttribute{Computed: true},
			"perm_publish_kafka":     schema.BoolAttribute{Computed: true},
			"perm_decaying":          schema.BoolAttribute{Computed: true},
			"perm_galaxy_editor":     schema.BoolAttribute{Computed: true},
			"default_role":           schema.BoolAttribute{Computed: true},
			"restricted_to_site_admin": schema.BoolAttribute{Computed: true},
			"enforce_rate_limit":     schema.BoolAttribute{Computed: true},
			"memory_limit":           schema.StringAttribute{Computed: true},
			"max_execution_time":     schema.StringAttribute{Computed: true},
			"rate_limit_count":       schema.StringAttribute{Computed: true},
		},
	}
}

func (d *roleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *roleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data roleResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := d.client.GetRole(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Reading MISP role failed", err.Error())
		return
	}
	out := roleToModel(got)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}
