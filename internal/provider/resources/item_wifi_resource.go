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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/planeopscc/terraform-provider-protonpass/internal/passcli"
)

var _ resource.Resource = &ItemWiFiResource{}
var _ resource.ResourceWithImportState = &ItemWiFiResource{}

type ItemWiFiResource struct {
	client *passcli.Client
}

type ItemWiFiResourceModel struct {
	ItemID        types.String `tfsdk:"item_id"`
	ShareID       types.String `tfsdk:"share_id"`
	Title         types.String `tfsdk:"title"`
	SSID          types.String `tfsdk:"ssid"`
	PasswordWO    types.String `tfsdk:"password_wo"`
	PasswordWOVer types.Int64  `tfsdk:"password_wo_version"`
	Security      types.String `tfsdk:"security"`
	CreateTime    types.String `tfsdk:"create_time"`
	ModifyTime    types.String `tfsdk:"modify_time"`
}

func NewItemWiFiResource() resource.Resource { return &ItemWiFiResource{} }

func (r *ItemWiFiResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_item_wifi"
}

func (r *ItemWiFiResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Proton Pass WiFi item.",
		Attributes: map[string]schema.Attribute{
			"item_id":             schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"share_id":            schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"title":               schema.StringAttribute{Required: true},
			"ssid":                schema.StringAttribute{Optional: true, MarkdownDescription: "WiFi network name (SSID)."},
			"password_wo":         schema.StringAttribute{Optional: true, WriteOnly: true, MarkdownDescription: "WiFi password (write-only)."},
			"password_wo_version": schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
			"security":            schema.StringAttribute{Optional: true, MarkdownDescription: "Security type (wpa2, wpa3, etc)."},
			"create_time":         schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"modify_time":         schema.StringAttribute{Computed: true},
		},
	}
}

func (r *ItemWiFiResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ItemWiFiResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ItemWiFiResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var pwWO types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("password_wo"), &pwWO)...)
	password := ""
	if !pwWO.IsNull() && !pwWO.IsUnknown() {
		password = pwWO.ValueString()
	}

	tflog.Debug(ctx, "creating item_wifi", map[string]interface{}{"title": data.Title.ValueString()})

	item, err := r.client.CreateItemWiFi(ctx, data.ShareID.ValueString(), data.Title.ValueString(), data.SSID.ValueString(), password, data.Security.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to create WiFi item", err.Error())
		return
	}
	data.ItemID = types.StringValue(item.ItemID)
	data.ShareID = types.StringValue(item.ShareID)
	data.CreateTime = types.StringValue(item.CreateTime)
	data.ModifyTime = types.StringValue(item.ModifyTime)
	if data.PasswordWOVer.IsNull() || data.PasswordWOVer.IsUnknown() {
		data.PasswordWOVer = types.Int64Value(0)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ItemWiFiResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ItemWiFiResourceModel
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
		resp.Diagnostics.AddError("Failed to read WiFi item", err.Error())
		return
	}
	data.Title = types.StringValue(item.Title)
	data.CreateTime = types.StringValue(item.CreateTime)
	data.ModifyTime = types.StringValue(item.ModifyTime)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ItemWiFiResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ItemWiFiResourceModel
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
			resp.Diagnostics.AddError("Failed to update WiFi item", err.Error())
			return
		}
	}
	item, err := r.client.ReadItemLogin(ctx, state.ItemID.ValueString(), state.ShareID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read WiFi after update", err.Error())
		return
	}
	plan.ItemID = state.ItemID
	plan.ShareID = state.ShareID
	plan.Title = types.StringValue(item.Title)
	plan.CreateTime = types.StringValue(item.CreateTime)
	plan.ModifyTime = types.StringValue(item.ModifyTime)
	if plan.PasswordWOVer.IsNull() || plan.PasswordWOVer.IsUnknown() {
		plan.PasswordWOVer = state.PasswordWOVer
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ItemWiFiResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ItemWiFiResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteItemLogin(ctx, data.ItemID.ValueString(), data.ShareID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to delete WiFi item", err.Error())
	}
}

func (r *ItemWiFiResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
