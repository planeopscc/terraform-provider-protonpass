// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/planeopscc/terraform-provider-protonpass/internal/passcli"
)

var _ datasource.DataSource = &ItemCreditCardDataSource{}

type ItemCreditCardDataSource struct {
	client *passcli.Client
}

type ItemCreditCardDataSourceModel struct {
	ItemID             types.String `tfsdk:"item_id"`
	ShareID            types.String `tfsdk:"share_id"`
	Title              types.String `tfsdk:"title"`
	CardholderName     types.String `tfsdk:"cardholder_name"`
	Number             types.String `tfsdk:"number"`
	VerificationNumber types.String `tfsdk:"verification_number"`
	ExpirationDate     types.String `tfsdk:"expiration_date"`
	PIN                types.String `tfsdk:"pin"`
	CreateTime         types.String `tfsdk:"create_time"`
	ModifyTime         types.String `tfsdk:"modify_time"`
}

func NewItemCreditCardDataSource() datasource.DataSource {
	return &ItemCreditCardDataSource{}
}

func (d *ItemCreditCardDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_item_credit_card"
}

func (d *ItemCreditCardDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a specific Proton Pass credit card item.",
		Attributes: map[string]schema.Attribute{
			"item_id": schema.StringAttribute{
				MarkdownDescription: "Item ID.",
				Required:            true,
			},
			"share_id": schema.StringAttribute{
				MarkdownDescription: "Share ID of the vault.",
				Required:            true,
			},
			"title":               schema.StringAttribute{Computed: true, MarkdownDescription: "Title of the item."},
			"cardholder_name":     schema.StringAttribute{Computed: true, MarkdownDescription: "Cardholder Name."},
			"number":              schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Card Number."},
			"verification_number": schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Verification Number (CVV)."},
			"expiration_date":     schema.StringAttribute{Computed: true, MarkdownDescription: "Expiration Date."},
			"pin":                 schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "PIN."},
			"create_time":         schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp."},
			"modify_time":         schema.StringAttribute{Computed: true, MarkdownDescription: "Last modification timestamp."},
		},
	}
}

func (d *ItemCreditCardDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*passcli.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *passcli.Client, got: %T.", req.ProviderData))
		return
	}
	d.client = client
}

func (d *ItemCreditCardDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ItemCreditCardDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	item, err := d.client.GetItem(ctx, data.ItemID.ValueString(), "", data.ShareID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get credit card item", err.Error())
		return
	}

	data.Title = types.StringValue(item.Content.Title)
	data.CreateTime = types.StringValue(item.CreateTime)
	data.ModifyTime = types.StringValue(item.ModifyTime)

	if item.Content.Content.CreditCard != nil {
		data.CardholderName = types.StringValue(item.Content.Content.CreditCard.CardholderName)
		data.Number = types.StringValue(item.Content.Content.CreditCard.Number)
		data.VerificationNumber = types.StringValue(item.Content.Content.CreditCard.VerificationNumber)
		data.ExpirationDate = types.StringValue(item.Content.Content.CreditCard.ExpirationDate)
		data.PIN = types.StringValue(item.Content.Content.CreditCard.PIN)
	} else {
		resp.Diagnostics.AddError("Invalid Item Type", "The specified item is not a Credit Card item.")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
