// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/planeopscc/terraform-provider-protonpass/internal/passcli"
)

var _ resource.Resource = &VaultResource{}
var _ resource.ResourceWithImportState = &VaultResource{}

type VaultResource struct {
	client *passcli.Client
}

type VaultResourceModel struct {
	ShareID types.String `tfsdk:"share_id"`
	VaultID types.String `tfsdk:"vault_id"`
	Name    types.String `tfsdk:"name"`
}

func NewVaultResource() resource.Resource {
	return &VaultResource{}
}

func (r *VaultResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault"
}

func (r *VaultResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Proton Pass vault.",
		Attributes: map[string]schema.Attribute{
			"share_id": schema.StringAttribute{
				MarkdownDescription: "Share ID of the vault (unique identifier).",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"vault_id": schema.StringAttribute{
				MarkdownDescription: "Internal vault ID.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the vault.",
				Required:            true,
			},
		},
	}
}

func (r *VaultResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*passcli.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *passcli.Client, got: %T.", req.ProviderData))
		return
	}
	r.client = client
}

func (r *VaultResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VaultResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "creating vault", map[string]interface{}{"name": data.Name.ValueString()})

	vault, err := r.client.CreateVault(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to create vault", err.Error())
		return
	}

	data.ShareID = types.StringValue(vault.ShareID)
	data.VaultID = types.StringValue(vault.VaultID)
	data.Name = types.StringValue(vault.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VaultResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VaultResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vault, err := r.client.ReadVault(ctx, data.ShareID.ValueString())
	if err != nil {
		if passcli.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read vault", err.Error())
		return
	}

	data.Name = types.StringValue(vault.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VaultResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state VaultResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Name.ValueString() != state.Name.ValueString() {
		err := r.client.UpdateVault(ctx, state.ShareID.ValueString(), plan.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to rename vault", err.Error())
			return
		}
	}

	plan.ShareID = state.ShareID
	plan.VaultID = state.VaultID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *VaultResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VaultResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.DeleteVault(ctx, data.ShareID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete vault", err.Error())
		return
	}
}

func (r *VaultResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("share_id"), req, resp)
}
