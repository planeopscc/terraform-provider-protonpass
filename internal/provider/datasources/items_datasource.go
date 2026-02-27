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

var _ datasource.DataSource = &ItemsDataSource{}

type ItemsDataSource struct {
	client *passcli.Client
}

type ItemsDataSourceModel struct {
	ShareID types.String    `tfsdk:"share_id"`
	Items   []ItemDataModel `tfsdk:"items"`
}

type ItemDataModel struct {
	ItemID     types.String `tfsdk:"item_id"`
	ShareID    types.String `tfsdk:"share_id"`
	Title      types.String `tfsdk:"title"`
	CreateTime types.String `tfsdk:"create_time"`
	ModifyTime types.String `tfsdk:"modify_time"`
}

func NewItemsDataSource() datasource.DataSource {
	return &ItemsDataSource{}
}

func (d *ItemsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_items"
}

func (d *ItemsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all items in a Proton Pass vault.",
		Attributes: map[string]schema.Attribute{
			"share_id": schema.StringAttribute{
				MarkdownDescription: "Share ID of the vault to list items from.",
				Required:            true,
			},
			"items": schema.ListNestedAttribute{
				MarkdownDescription: "List of items in the vault.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"item_id":     schema.StringAttribute{Computed: true, MarkdownDescription: "Unique identifier of the item."},
						"share_id":    schema.StringAttribute{Computed: true, MarkdownDescription: "Share ID of the vault."},
						"title":       schema.StringAttribute{Computed: true, MarkdownDescription: "Title of the item."},
						"create_time": schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp."},
						"modify_time": schema.StringAttribute{Computed: true, MarkdownDescription: "Last modification timestamp."},
					},
				},
			},
		},
	}
}

func (d *ItemsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ItemsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ItemsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, err := d.client.ListItemsInVault(ctx, data.ShareID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to list items", err.Error())
		return
	}

	data.Items = make([]ItemDataModel, len(items))
	for i, item := range items {
		data.Items[i] = ItemDataModel{
			ItemID:     types.StringValue(item.ItemID),
			ShareID:    types.StringValue(item.ShareID),
			Title:      types.StringValue(item.Title),
			CreateTime: types.StringValue(item.CreateTime),
			ModifyTime: types.StringValue(item.ModifyTime),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
