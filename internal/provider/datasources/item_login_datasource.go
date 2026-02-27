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

var _ datasource.DataSource = &ItemLoginDataSource{}

type ItemLoginDataSource struct {
	client *passcli.Client
}

type ItemLoginDataSourceModel struct {
	ItemID     types.String `tfsdk:"item_id"`
	ShareID    types.String `tfsdk:"share_id"`
	Title      types.String `tfsdk:"title"`
	Username   types.String `tfsdk:"username"`
	Email      types.String `tfsdk:"email"`
	Password   types.String `tfsdk:"password"`
	URLs       []string     `tfsdk:"urls"`
	CreateTime types.String `tfsdk:"create_time"`
	ModifyTime types.String `tfsdk:"modify_time"`
}

func NewItemLoginDataSource() datasource.DataSource {
	return &ItemLoginDataSource{}
}

func (d *ItemLoginDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_item_login"
}

func (d *ItemLoginDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a specific Proton Pass login item.",
		Attributes: map[string]schema.Attribute{
			"item_id": schema.StringAttribute{
				MarkdownDescription: "Item ID.",
				Required:            true,
			},
			"share_id": schema.StringAttribute{
				MarkdownDescription: "Share ID of the vault.",
				Required:            true,
			},
			"title":       schema.StringAttribute{Computed: true, MarkdownDescription: "Title of the login item."},
			"username":    schema.StringAttribute{Computed: true, MarkdownDescription: "Username."},
			"email":       schema.StringAttribute{Computed: true, MarkdownDescription: "Email."},
			"password":    schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Password."},
			"urls":        schema.ListAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "URLs."},
			"create_time": schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp."},
			"modify_time": schema.StringAttribute{Computed: true, MarkdownDescription: "Last modification timestamp."},
		},
	}
}

func (d *ItemLoginDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ItemLoginDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ItemLoginDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	item, err := d.client.GetItem(ctx, data.ItemID.ValueString(), "", data.ShareID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get login item", err.Error())
		return
	}

	data.Title = types.StringValue(item.Content.Title)
	data.CreateTime = types.StringValue(item.CreateTime)
	data.ModifyTime = types.StringValue(item.ModifyTime)

	if item.Content.Content.Login != nil {
		data.Username = types.StringValue(item.Content.Content.Login.Username)
		data.Email = types.StringValue(item.Content.Content.Login.Email)
		data.Password = types.StringValue(item.Content.Content.Login.Password)
		data.URLs = item.Content.Content.Login.URLs
	} else {
		resp.Diagnostics.AddError("Invalid Item Type", "The specified item is not a Login item.")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
