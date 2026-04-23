// Package provider implements the MISP Terraform provider.
package provider

import (
	"context"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Scriptception/terraform-provider-misp/internal/client"
)

// Ensure MISPProvider satisfies various provider interfaces.
var _ provider.Provider = (*MISPProvider)(nil)

// MISPProvider is the Terraform provider implementation.
type MISPProvider struct {
	version string
}

// MISPProviderModel describes the provider data model.
type MISPProviderModel struct {
	URL      types.String `tfsdk:"url"`
	APIKey   types.String `tfsdk:"api_key"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

// New returns a provider constructor bound to a version string set at build time.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &MISPProvider{version: version}
	}
}

// Metadata returns the provider type name ("misp") and its build version.
func (p *MISPProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "misp"
	resp.Version = p.version
}

// Schema defines the provider-block configuration schema.
func (p *MISPProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage a MISP (Malware Information Sharing Platform) instance as code.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "Base URL of the MISP instance (e.g. `https://misp.example.com`). May also be set via the `MISP_URL` environment variable.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "MISP API key. May also be set via the `MISP_API_KEY` environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"insecure": schema.BoolAttribute{
				Description: "Skip TLS certificate verification. May also be set via `MISP_INSECURE`. Disabled by default.",
				Optional:    true,
			},
		},
	}
}

// Configure builds the MISP client from provider config (or MISP_* env vars)
// and stashes it on the response so resources and data sources can reach it.
func (p *MISPProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data MISPProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := firstNonEmpty(data.URL.ValueString(), os.Getenv("MISP_URL"))
	apiKey := firstNonEmpty(data.APIKey.ValueString(), os.Getenv("MISP_API_KEY"))

	insecure := data.Insecure.ValueBool()
	if data.Insecure.IsNull() {
		if v, err := strconv.ParseBool(os.Getenv("MISP_INSECURE")); err == nil {
			insecure = v
		}
	}

	if url == "" {
		resp.Diagnostics.AddAttributeError(
			pathRoot("url"),
			"Missing MISP URL",
			"Set the `url` attribute or the MISP_URL environment variable.",
		)
	}
	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			pathRoot("api_key"),
			"Missing MISP API key",
			"Set the `api_key` attribute or the MISP_API_KEY environment variable.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	c, err := client.New(client.Config{
		URL:       url,
		APIKey:    apiKey,
		Insecure:  insecure,
		UserAgent: "terraform-provider-misp/" + p.version,
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create MISP client", err.Error())
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

// Resources returns the set of resources the provider supports.
func (p *MISPProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewOrganisationResource,
		NewTagResource,
		NewSharingGroupResource,
		NewSharingGroupMemberResource,
		NewSharingGroupServerResource,
		NewUserResource,
		NewTaxonomyResource,
		NewFeedResource,
		NewServerResource,
		NewRoleResource,
		NewWarninglistResource,
		NewNoticelistResource,
		NewSettingResource,
		NewGalaxyClusterResource,
	}
}

// DataSources returns the set of data sources the provider supports.
func (p *MISPProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewOrganisationDataSource,
		NewTagDataSource,
		NewSharingGroupDataSource,
		NewUserDataSource,
		NewTaxonomyDataSource,
		NewFeedDataSource,
		NewServerDataSource,
		NewRoleDataSource,
		NewWarninglistDataSource,
		NewNoticelistDataSource,
		NewSettingDataSource,
		NewGalaxyClusterDataSource,
	}
}

// Functions returns provider-defined functions; none are defined yet.
func (p *MISPProvider) Functions(_ context.Context) []func() function.Function {
	return nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
