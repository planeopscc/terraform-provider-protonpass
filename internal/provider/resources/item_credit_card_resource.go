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

var _ resource.Resource = &ItemCreditCardResource{}
var _ resource.ResourceWithImportState = &ItemCreditCardResource{}

type ItemCreditCardResource struct {
	client *passcli.Client
}

type ItemCreditCardResourceModel struct {
	ItemID          types.String `tfsdk:"item_id"`
	ShareID         types.String `tfsdk:"share_id"`
	Title           types.String `tfsdk:"title"`
	CardholderName  types.String `tfsdk:"cardholder_name"`
	CardNumberWO    types.String `tfsdk:"card_number_wo"`
	CardNumberWOVer types.Int64  `tfsdk:"card_number_wo_version"`
	CvvWO           types.String `tfsdk:"cvv_wo"`
	CvvWOVer        types.Int64  `tfsdk:"cvv_wo_version"`
	ExpirationDate  types.String `tfsdk:"expiration_date"`
	PinWO           types.String `tfsdk:"pin_wo"`
	PinWOVer        types.Int64  `tfsdk:"pin_wo_version"`
	CreateTime      types.String `tfsdk:"create_time"`
	ModifyTime      types.String `tfsdk:"modify_time"`
}

func NewItemCreditCardResource() resource.Resource { return &ItemCreditCardResource{} }

func (r *ItemCreditCardResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_item_credit_card"
}

func (r *ItemCreditCardResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Proton Pass credit card item.",
		Attributes: map[string]schema.Attribute{
			"item_id":                schema.StringAttribute{MarkdownDescription: "Unique identifier of the item.", Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"share_id":               schema.StringAttribute{MarkdownDescription: "Share ID of the vault.", Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"title":                  schema.StringAttribute{MarkdownDescription: "Title of the credit card item.", Required: true},
			"cardholder_name":        schema.StringAttribute{MarkdownDescription: "Name of the cardholder.", Optional: true},
			"card_number_wo":         schema.StringAttribute{MarkdownDescription: "Card number (write-only, never stored in state).", Optional: true, WriteOnly: true},
			"card_number_wo_version": schema.Int64Attribute{MarkdownDescription: "Increment to trigger card number update.", Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
			"cvv_wo":                 schema.StringAttribute{MarkdownDescription: "CVV/CVC code (write-only, never stored in state).", Optional: true, WriteOnly: true},
			"cvv_wo_version":         schema.Int64Attribute{MarkdownDescription: "Increment to trigger CVV update.", Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
			"expiration_date":        schema.StringAttribute{MarkdownDescription: "Expiration date in YYYY-MM format.", Optional: true},
			"pin_wo":                 schema.StringAttribute{MarkdownDescription: "Card PIN (write-only, never stored in state).", Optional: true, WriteOnly: true},
			"pin_wo_version":         schema.Int64Attribute{MarkdownDescription: "Increment to trigger PIN update.", Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
			"create_time":            schema.StringAttribute{MarkdownDescription: "Creation timestamp.", Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"modify_time":            schema.StringAttribute{MarkdownDescription: "Last modification timestamp.", Computed: true},
		},
	}
}

func (r *ItemCreditCardResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ItemCreditCardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ItemCreditCardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readWO := func(attr string) string {
		var v types.String
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(attr), &v)...)
		if !v.IsNull() && !v.IsUnknown() {
			return v.ValueString()
		}
		return ""
	}

	tflog.Debug(ctx, "creating item_credit_card", map[string]interface{}{"title": data.Title.ValueString()})

	item, err := r.client.CreateItemCreditCard(ctx,
		data.ShareID.ValueString(), data.Title.ValueString(),
		data.CardholderName.ValueString(), readWO("card_number_wo"),
		readWO("cvv_wo"), data.ExpirationDate.ValueString(), readWO("pin_wo"))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create credit card item", err.Error())
		return
	}

	data.ItemID = types.StringValue(item.ItemID)
	data.ShareID = types.StringValue(item.ShareID)
	data.CreateTime = types.StringValue(item.CreateTime)
	data.ModifyTime = types.StringValue(item.ModifyTime)
	if data.CardNumberWOVer.IsNull() || data.CardNumberWOVer.IsUnknown() {
		data.CardNumberWOVer = types.Int64Value(0)
	}
	if data.CvvWOVer.IsNull() || data.CvvWOVer.IsUnknown() {
		data.CvvWOVer = types.Int64Value(0)
	}
	if data.PinWOVer.IsNull() || data.PinWOVer.IsUnknown() {
		data.PinWOVer = types.Int64Value(0)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ItemCreditCardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ItemCreditCardResourceModel
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
		resp.Diagnostics.AddError("Failed to read credit card item", err.Error())
		return
	}
	data.Title = types.StringValue(item.Title)
	data.CreateTime = types.StringValue(item.CreateTime)
	data.ModifyTime = types.StringValue(item.ModifyTime)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ItemCreditCardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ItemCreditCardResourceModel
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
			resp.Diagnostics.AddError("Failed to update credit card item", err.Error())
			return
		}
	}
	item, err := r.client.ReadItemLogin(ctx, state.ItemID.ValueString(), state.ShareID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read credit card after update", err.Error())
		return
	}
	plan.ItemID = state.ItemID
	plan.ShareID = state.ShareID
	plan.Title = types.StringValue(item.Title)
	plan.CreateTime = types.StringValue(item.CreateTime)
	plan.ModifyTime = types.StringValue(item.ModifyTime)
	if plan.CardNumberWOVer.IsNull() || plan.CardNumberWOVer.IsUnknown() {
		plan.CardNumberWOVer = state.CardNumberWOVer
	}
	if plan.CvvWOVer.IsNull() || plan.CvvWOVer.IsUnknown() {
		plan.CvvWOVer = state.CvvWOVer
	}
	if plan.PinWOVer.IsNull() || plan.PinWOVer.IsUnknown() {
		plan.PinWOVer = state.PinWOVer
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ItemCreditCardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ItemCreditCardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteItemLogin(ctx, data.ItemID.ValueString(), data.ShareID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to delete credit card item", err.Error())
	}
}

func (r *ItemCreditCardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
