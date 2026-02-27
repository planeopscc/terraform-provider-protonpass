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

var _ resource.Resource = &ItemSSHKeyResource{}
var _ resource.ResourceWithImportState = &ItemSSHKeyResource{}

type ItemSSHKeyResource struct {
	client *passcli.Client
}

type ItemSSHKeyResourceModel struct {
	ItemID     types.String `tfsdk:"item_id"`
	ShareID    types.String `tfsdk:"share_id"`
	Title      types.String `tfsdk:"title"`
	KeyType    types.String `tfsdk:"key_type"`
	Comment    types.String `tfsdk:"comment"`
	CreateTime types.String `tfsdk:"create_time"`
	ModifyTime types.String `tfsdk:"modify_time"`
}

func NewItemSSHKeyResource() resource.Resource { return &ItemSSHKeyResource{} }

func (r *ItemSSHKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_item_ssh_key"
}

func (r *ItemSSHKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Proton Pass SSH key item. Generates a new SSH key pair.",
		Attributes: map[string]schema.Attribute{
			"item_id":     schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"share_id":    schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"title":       schema.StringAttribute{Required: true},
			"key_type":    schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Type of SSH key: ed25519, rsa2048, rsa4096. Default: ed25519.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"comment":     schema.StringAttribute{Optional: true, MarkdownDescription: "Comment for the SSH key."},
			"create_time": schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"modify_time": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *ItemSSHKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ItemSSHKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ItemSSHKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	keyType := data.KeyType.ValueString()
	if keyType == "" {
		keyType = "ed25519"
	}

	tflog.Debug(ctx, "creating item_ssh_key", map[string]interface{}{"title": data.Title.ValueString(), "key_type": keyType})

	item, err := r.client.CreateItemSSHKey(ctx, data.ShareID.ValueString(), data.Title.ValueString(), keyType, data.Comment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to create SSH key item", err.Error())
		return
	}
	data.ItemID = types.StringValue(item.ItemID)
	data.ShareID = types.StringValue(item.ShareID)
	data.Title = types.StringValue(item.Title)
	data.KeyType = types.StringValue(keyType)
	data.CreateTime = types.StringValue(item.CreateTime)
	data.ModifyTime = types.StringValue(item.ModifyTime)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ItemSSHKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ItemSSHKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	item, err := r.client.ReadItemLogin(ctx, data.ItemID.ValueString(), data.ShareID.ValueString())
	if err != nil {
		if passcli.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read SSH key item", err.Error())
		return
	}
	data.Title = types.StringValue(item.Title)
	data.CreateTime = types.StringValue(item.CreateTime)
	data.ModifyTime = types.StringValue(item.ModifyTime)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ItemSSHKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ItemSSHKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fields := map[string]string{}
	if plan.Title.ValueString() != state.Title.ValueString() {
		fields["title"] = plan.Title.ValueString()
	}
	if len(fields) > 0 {
		if err := r.client.UpdateItemLogin(ctx, state.ItemID.ValueString(), state.ShareID.ValueString(), fields); err != nil {
			resp.Diagnostics.AddError("Failed to update SSH key item", err.Error())
			return
		}
	}
	item, err := r.client.ReadItemLogin(ctx, state.ItemID.ValueString(), state.ShareID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read SSH key after update", err.Error())
		return
	}
	plan.ItemID = state.ItemID
	plan.ShareID = state.ShareID
	plan.KeyType = state.KeyType
	plan.Title = types.StringValue(item.Title)
	plan.CreateTime = types.StringValue(item.CreateTime)
	plan.ModifyTime = types.StringValue(item.ModifyTime)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ItemSSHKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ItemSSHKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteItemLogin(ctx, data.ItemID.ValueString(), data.ShareID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to delete SSH key item", err.Error())
	}
}

func (r *ItemSSHKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: share_id:item_id
	idParts := strings.Split(req.ID, ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: share_id:item_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("share_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("item_id"), idParts[1])...)
}
