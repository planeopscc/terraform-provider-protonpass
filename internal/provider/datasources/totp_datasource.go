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

var _ datasource.DataSource = &TotpDataSource{}

func NewTotpDataSource() datasource.DataSource {
	return &TotpDataSource{}
}

type TotpDataSource struct {
	client *passcli.Client
}

type TotpDataSourceModel struct {
	ShareID types.String `tfsdk:"share_id"`
	ItemID  types.String `tfsdk:"item_id"`
	Code    types.String `tfsdk:"code"`
	ID      types.String `tfsdk:"id"`
}

func (d *TotpDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_totp"
}

func (d *TotpDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves the current TOTP (Time-Based One-Time Password) code for a specific item.",
		Attributes: map[string]schema.Attribute{
			"share_id": schema.StringAttribute{
				MarkdownDescription: "The Share ID of the vault containing the item.",
				Required:            true,
			},
			"item_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the item containing the TOTP secret.",
				Required:            true,
			},
			"code": schema.StringAttribute{
				MarkdownDescription: "The generated TOTP code.",
				Computed:            true,
				Sensitive:           true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The standard Terraform identifier.",
				Computed:            true,
			},
		},
	}
}

func (d *TotpDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*passcli.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *passcli.Client, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *TotpDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TotpDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	totp, err := d.client.GetItemTOTP(ctx, data.ItemID.ValueString(), data.ShareID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read TOTP code", err.Error())
		return
	}

	data.Code = types.StringValue(totp.Code)
	// TF Data Sources usually need an ID. Combining shareID and itemID makes a unique identifiable string
	data.ID = types.StringValue(fmt.Sprintf("%s:%s", data.ShareID.ValueString(), data.ItemID.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
