// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/planeopscc/terraform-provider-protonpass/internal/passcli"
)

var _ resource.Resource = &VaultMemberResource{}
var _ resource.ResourceWithImportState = &VaultMemberResource{}

func NewVaultMemberResource() resource.Resource {
	return &VaultMemberResource{}
}

type VaultMemberResource struct {
	client *passcli.Client
}

type VaultMemberResourceModel struct {
	ShareID       types.String `tfsdk:"share_id"`
	Email         types.String `tfsdk:"email"`
	Role          types.String `tfsdk:"role"`
	MemberShareID types.String `tfsdk:"member_share_id"`
}

func (r *VaultMemberResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_member"
}

func (r *VaultMemberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a member's access to a Proton Pass Vault.",
		Attributes: map[string]schema.Attribute{
			"share_id": schema.StringAttribute{
				MarkdownDescription: "The Share ID of the vault.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The email address of the user to invite.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "The role of the user within the vault. Allowed values: `viewer`, `editor`, `manager`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("viewer", "editor", "manager"),
				},
			},
			"member_share_id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier for this member's access to the vault.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *VaultMemberResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VaultMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VaultMemberResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "creating vault_member", map[string]interface{}{
		"share_id": data.ShareID.ValueString(),
		"email":    data.Email.ValueString(),
		"role":     data.Role.ValueString(),
	})

	member, err := r.client.AddVaultMember(ctx, data.ShareID.ValueString(), data.Email.ValueString(), data.Role.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to add vault member", err.Error())
		return
	}

	data.MemberShareID = types.StringValue(member.MemberShareID)
	// Proton Pass CLI title-cases roles (e.g. "Viewer"). We lower-case it for TF state matching
	data.Role = types.StringValue(strings.ToLower(member.Role))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VaultMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VaultMemberResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	member, err := r.client.ReadVaultMember(ctx, data.ShareID.ValueString(), data.MemberShareID.ValueString(), "")
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			tflog.Warn(ctx, "vault_member not found, removing from state", map[string]interface{}{
				"member_share_id": data.MemberShareID.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read vault member", err.Error())
		return
	}

	data.Email = types.StringValue(member.Email)
	data.MemberShareID = types.StringValue(member.MemberShareID)
	if member.Role != "unknown" {
		data.Role = types.StringValue(strings.ToLower(member.Role))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VaultMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state VaultMemberResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Role.ValueString() != state.Role.ValueString() {
		err := r.client.UpdateVaultMemberRole(ctx, state.ShareID.ValueString(), state.MemberShareID.ValueString(), plan.Role.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to update vault member role", err.Error())
			return
		}
	}

	plan.MemberShareID = state.MemberShareID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *VaultMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VaultMemberResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RemoveVaultMember(ctx, data.ShareID.ValueString(), data.MemberShareID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to remove vault member", err.Error())
		return
	}
}

func (r *VaultMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: share_id:member_share_id
	idParts := strings.Split(req.ID, ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: share_id:member_share_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("share_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("member_share_id"), idParts[1])...)
}
