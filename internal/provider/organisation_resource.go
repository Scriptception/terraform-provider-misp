package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Scriptception/terraform-provider-misp/internal/client"
)

var (
	_ resource.Resource                = (*organisationResource)(nil)
	_ resource.ResourceWithConfigure   = (*organisationResource)(nil)
	_ resource.ResourceWithImportState = (*organisationResource)(nil)
)

// NewOrganisationResource constructs a misp_organisation resource.
func NewOrganisationResource() resource.Resource {
	return &organisationResource{}
}

type organisationResource struct {
	client *client.Client
}

type organisationResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	UUID               types.String `tfsdk:"uuid"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	Type               types.String `tfsdk:"type"`
	Nationality        types.String `tfsdk:"nationality"`
	Sector             types.String `tfsdk:"sector"`
	Contacts           types.String `tfsdk:"contacts"`
	Local              types.Bool   `tfsdk:"local"`
	RestrictedToDomain types.List   `tfsdk:"restricted_to_domain"`
}

func (r *organisationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organisation"
}

func (r *organisationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A MISP organisation. Organisations own events and attributes and are the primary unit of access control.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Numeric MISP identifier.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uuid": schema.StringAttribute{
				Description: "Stable UUID assigned by MISP.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Organisation name. Must be unique within the MISP instance.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Free-form description.",
				Optional:    true,
			},
			"type": schema.StringAttribute{
				Description: "Organisation type label (e.g. `ADMIN`, `CSIRT`).",
				Optional:    true,
			},
			"nationality": schema.StringAttribute{
				Description: "Country code.",
				Optional:    true,
			},
			"sector": schema.StringAttribute{
				Description: "Industry sector.",
				Optional:    true,
			},
			"contacts": schema.StringAttribute{
				Description: "Contact details.",
				Optional:    true,
			},
			"local": schema.BoolAttribute{
				Description: "Whether this organisation is local to the instance (true) or known only by reference (false).",
				Optional:    true,
				Computed:    true,
			},
			"restricted_to_domain": schema.ListAttribute{
				Description: "Restrict organisation membership to users with email addresses on these domains.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *organisationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("expected *client.Client, got %T", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *organisationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organisationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateOrganisation(ctx, orgFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Creating MISP organisation failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, orgToModel(created))...)
}

func (r *organisationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organisationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	got, err := r.client.GetOrganisation(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Reading MISP organisation failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, orgToModel(got))...)
}

func (r *organisationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan organisationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state organisationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateOrganisation(ctx, state.ID.ValueString(), orgFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Updating MISP organisation failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, orgToModel(updated))...)
}

func (r *organisationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organisationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteOrganisation(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Deleting MISP organisation failed", err.Error())
	}
}

func (r *organisationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func orgFromModel(m organisationResourceModel) client.Organisation {
	var domains []string
	if !m.RestrictedToDomain.IsNull() && !m.RestrictedToDomain.IsUnknown() {
		m.RestrictedToDomain.ElementsAs(context.Background(), &domains, false)
	}
	return client.Organisation{
		Name:               m.Name.ValueString(),
		Description:        m.Description.ValueString(),
		Type:               m.Type.ValueString(),
		Nationality:        m.Nationality.ValueString(),
		Sector:             m.Sector.ValueString(),
		Contacts:           m.Contacts.ValueString(),
		Local:              m.Local.ValueBool(),
		RestrictedToDomain: client.DomainList(domains),
	}
}

func orgToModel(o *client.Organisation) organisationResourceModel {
	m := organisationResourceModel{
		ID:    types.StringValue(o.ID),
		UUID:  types.StringValue(o.UUID),
		Name:  types.StringValue(o.Name),
		Local: types.BoolValue(o.Local),
	}
	m.Description = stringOrNull(o.Description)
	m.Type = stringOrNull(o.Type)
	m.Nationality = stringOrNull(o.Nationality)
	m.Sector = stringOrNull(o.Sector)
	m.Contacts = stringOrNull(o.Contacts)

	domainVals := make([]attr.Value, 0, len(o.RestrictedToDomain))
	for _, d := range o.RestrictedToDomain {
		domainVals = append(domainVals, types.StringValue(d))
	}
	// For a []string source, ListValue never returns diagnostics.
	rtd, _ := types.ListValue(types.StringType, domainVals)
	m.RestrictedToDomain = rtd

	return m
}

func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
