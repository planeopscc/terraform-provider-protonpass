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

var _ datasource.DataSource = &ItemWifiDataSource{}

type ItemWifiDataSource struct {
	client *passcli.Client
}

type ItemWifiDataSourceModel struct {
	ItemID     types.String `tfsdk:"item_id"`
	ShareID    types.String `tfsdk:"share_id"`
	Title      types.String `tfsdk:"title"`
	SSID       types.String `tfsdk:"ssid"`
	Password   types.String `tfsdk:"password"`
	Security   types.String `tfsdk:"security"`
	CreateTime types.String `tfsdk:"create_time"`
	ModifyTime types.String `tfsdk:"modify_time"`
}

func NewItemWifiDataSource() datasource.DataSource {
	return &ItemWifiDataSource{}
}

func (d *ItemWifiDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_item_wifi"
}

func (d *ItemWifiDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a specific Proton Pass WiFi item.",
		Attributes: map[string]schema.Attribute{
			"item_id": schema.StringAttribute{
				MarkdownDescription: "Item ID.",
				Required:            true,
			},
			"share_id": schema.StringAttribute{
				MarkdownDescription: "Share ID of the vault.",
				Required:            true,
			},
			"title":       schema.StringAttribute{Computed: true, MarkdownDescription: "Title of the item."},
			"ssid":        schema.StringAttribute{Computed: true, MarkdownDescription: "SSID."},
			"password":    schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Password."},
			"security":    schema.StringAttribute{Computed: true, MarkdownDescription: "Security protocol."},
			"create_time": schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp."},
			"modify_time": schema.StringAttribute{Computed: true, MarkdownDescription: "Last modification timestamp."},
		},
	}
}

func (d *ItemWifiDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ItemWifiDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ItemWifiDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	item, err := d.client.GetItem(ctx, data.ItemID.ValueString(), "", data.ShareID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get WiFi item", err.Error())
		return
	}

	data.Title = types.StringValue(item.Content.Title)
	data.CreateTime = types.StringValue(item.CreateTime)
	data.ModifyTime = types.StringValue(item.ModifyTime)

	if item.Content.Content.Wifi != nil {
		data.SSID = types.StringValue(item.Content.Content.Wifi.SSID)
		data.Password = types.StringValue(item.Content.Content.Wifi.Password)
		data.Security = types.StringValue(item.Content.Content.Wifi.Security)
	} else {
		resp.Diagnostics.AddError("Invalid Item Type", "The specified item is not a WiFi item.")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
