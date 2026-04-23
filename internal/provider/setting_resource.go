package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Scriptception/terraform-provider-misp/internal/client"
)

var (
	_ resource.Resource                = (*settingResource)(nil)
	_ resource.ResourceWithConfigure   = (*settingResource)(nil)
	_ resource.ResourceWithImportState = (*settingResource)(nil)
)

// NewSettingResource constructs a misp_setting resource.
//
// MISP settings are pre-defined on the server — they cannot be created or
// deleted via the API. This resource adopts an existing setting by name and
// manages its value. Destroying the resource is a no-op on the MISP side:
// it removes Terraform's tracking only; the value on the MISP instance is
// unchanged.
func NewSettingResource() resource.Resource {
	return &settingResource{}
}

type settingResource struct {
	client *client.Client
}

type settingResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Value       types.String `tfsdk:"value"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
	Level       types.Int64  `tfsdk:"level"`
}

func (r *settingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_setting"
}

func (r *settingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manage the value of a pre-existing MISP server setting.

MISP settings are built into the platform — they cannot be created or deleted
via the API. This resource **adopts** an existing setting identified by ` + "`name`" + ` and
manages its ` + "`value`" + `.

**Important — value is always a string.** MISP's API accepts the value as a
plain JSON string regardless of the setting's underlying type. For boolean
settings pass ` + `"true"` + ` or ` + `"false"` + `; for numeric settings pass the number as
a string, e.g. ` + `"5000"` + `. MISP coerces the string to the appropriate type
internally. The ` + "`type`" + ` attribute shows the underlying MISP type.

**Important — destroy is a no-op.** MISP has no API to delete or reset a
setting to its default. Running ` + "`terraform destroy`" + ` (or removing this resource
from your configuration) removes the resource from Terraform state only — the
value on the MISP instance is left unchanged. A warning diagnostic is emitted
to make this explicit. If you need to restore a setting to its default, do so
manually in the MISP administration UI.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Mirrors the setting name; used as the Terraform resource identity.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Dotted MISP setting name (e.g. `MISP.baseurl`). Replaces the resource if changed.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Description: "Desired value for the setting. Always pass as a string — MISP coerces it internally. For booleans use `\"true\"` or `\"false\"`; for numerics use the number as a string, e.g. `\"5000\"`.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "MISP-reported type of the setting (e.g. `string`, `boolean`, `numeric`).",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of the setting as reported by MISP.",
				Computed:    true,
			},
			"level": schema.Int64Attribute{
				Description: "MISP criticality level for the setting (0 = critical, 1 = recommended, 2 = optional).",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *settingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *settingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan settingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UpdateSetting(ctx, plan.Name.ValueString(), plan.Value.ValueString()); err != nil {
		resp.Diagnostics.AddError("Updating MISP setting failed", err.Error())
		return
	}

	// Re-read to capture the authoritative state after the edit.
	got, err := r.client.GetSetting(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Reading MISP setting failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, settingToModel(got))...)
}

func (r *settingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state settingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetSetting(ctx, state.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Reading MISP setting failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, settingToModel(got))...)
}

func (r *settingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan settingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UpdateSetting(ctx, plan.Name.ValueString(), plan.Value.ValueString()); err != nil {
		resp.Diagnostics.AddError("Updating MISP setting failed", err.Error())
		return
	}

	got, err := r.client.GetSetting(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Reading MISP setting failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, settingToModel(got))...)
}

// Delete is a deliberate no-op. MISP provides no API endpoint to delete or
// reset a setting to its default value. Destroying this resource removes it
// from Terraform state only — the value on the MISP instance is unchanged.
func (r *settingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state settingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.AddWarning(
		"MISP setting not reverted on destroy",
		fmt.Sprintf(
			"misp_setting %q has been removed from Terraform state, but its value on the MISP instance has NOT been changed. "+
				"MISP provides no API to delete or reset a setting to its default. "+
				"If you need to restore the original value, do so manually in the MISP administration UI.",
			state.Name.ValueString(),
		),
	)
}

// ImportState imports a setting by its dotted name (e.g. "MISP.baseurl").
// After import the Read function populates all computed attributes.
func (r *settingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The import ID is the setting name. We store it in both "id" and "name"
	// so that the subsequent Read (which keys off "name") works correctly.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}

func settingToModel(s *client.Setting) settingResourceModel {
	return settingResourceModel{
		// id mirrors name — it is the stable identifier for this resource.
		ID:          types.StringValue(s.Name),
		Name:        types.StringValue(s.Name),
		Value:       types.StringValue(s.Value.String()),
		Type:        types.StringValue(s.Type),
		Description: types.StringValue(s.Description),
		Level:       types.Int64Value(int64(s.Level)),
	}
}
