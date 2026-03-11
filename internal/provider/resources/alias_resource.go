// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/planeopscc/terraform-provider-protonpass/internal/passcli"
)

var _ resource.Resource = &AliasResource{}
var _ resource.ResourceWithImportState = &AliasResource{}

// NewAliasResource creates a new protonpass_alias resource.
func NewAliasResource() resource.Resource {
	return &AliasResource{}
}

// AliasResource manages a Proton Pass Hide-My-Email alias.
type AliasResource struct {
	client *passcli.Client
}

// AliasResourceModel describes the resource data model.
type AliasResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	ShareID            types.String `tfsdk:"share_id"`
	Prefix             types.String `tfsdk:"prefix"`
	Alias              types.String `tfsdk:"alias"`
	DestroyPermanently types.Bool   `tfsdk:"destroy_permanently"`
}

// Metadata returns the resource type name.
func (r *AliasResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alias"
}

// Schema defines the schema for the resource.
func (r *AliasResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a Hide-My-Email Alias in Proton Pass.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the alias item.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"share_id": schema.StringAttribute{
				MarkdownDescription: "The Share ID of the vault where the alias is created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"prefix": schema.StringAttribute{
				MarkdownDescription: "The prefix of the email alias (e.g., `my_prefix` results in `my_prefix.random@passmail.net`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"alias": schema.StringAttribute{
				MarkdownDescription: "The fully generated email address alias provided by Proton Pass.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"destroy_permanently": schema.BoolAttribute{
				MarkdownDescription: "If true, the alias is permanently deleted on destroy instead of moved to trash.",
				Optional:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *AliasResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*passcli.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *passcli.Client, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *AliasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AliasResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "creating alias", map[string]interface{}{
		"share_id": data.ShareID.ValueString(),
		"prefix":   data.Prefix.ValueString(),
	})

	alias, err := r.client.CreateAlias(ctx, data.ShareID.ValueString(), data.Prefix.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to create alias", err.Error())
		return
	}

	data.ID = types.StringValue(alias.ID)
	data.Alias = types.StringValue(alias.Alias)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *AliasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AliasResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	item, err := r.client.ReadItem(ctx, data.ID.ValueString(), data.ShareID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "Could not find item") {
			tflog.Warn(ctx, "alias not found, removing from state", map[string]interface{}{
				"id": data.ID.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}
		// Aliases may return 422 if read too quickly. Log and retain current state.
		tflog.Warn(ctx, "failed to read alias, retaining state", map[string]interface{}{"error": err.Error()})
		return
	}

	// An alias is internally a login item; the username field holds the generated alias address.
	if item.Username != "" {
		data.Alias = types.StringValue(item.Username)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *AliasResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state AliasResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// All mutable attributes use RequiresReplace; no in-place update is needed.
	plan.ID = state.ID
	plan.Alias = state.Alias
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *AliasResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AliasResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteItem(ctx, data.ID.ValueString(), data.ShareID.ValueString(), data.DestroyPermanently.ValueBool())
	if err != nil && !strings.Contains(err.Error(), "Could not find item") {
		resp.Diagnostics.AddError("Failed to delete alias", err.Error())
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *AliasResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected format: share_id:item_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("share_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}
