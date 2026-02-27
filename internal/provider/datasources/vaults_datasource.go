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

var _ datasource.DataSource = &VaultsDataSource{}

type VaultsDataSource struct {
	client *passcli.Client
}

type VaultsDataSourceModel struct {
	Vaults []VaultDataModel `tfsdk:"vaults"`
}

type VaultDataModel struct {
	ShareID types.String `tfsdk:"share_id"`
	VaultID types.String `tfsdk:"vault_id"`
	Name    types.String `tfsdk:"name"`
}

func NewVaultsDataSource() datasource.DataSource { return &VaultsDataSource{} }

func (d *VaultsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vaults"
}

func (d *VaultsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Proton Pass vaults.",
		Attributes: map[string]schema.Attribute{
			"vaults": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"share_id": schema.StringAttribute{Computed: true},
						"vault_id": schema.StringAttribute{Computed: true},
						"name":     schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *VaultsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VaultsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	vaults, err := d.client.ListVaults(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list vaults", err.Error())
		return
	}
	var data VaultsDataSourceModel
	data.Vaults = make([]VaultDataModel, len(vaults))
	for i, v := range vaults {
		data.Vaults[i] = VaultDataModel{
			ShareID: types.StringValue(v.ShareID),
			VaultID: types.StringValue(v.VaultID),
			Name:    types.StringValue(v.Name),
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
