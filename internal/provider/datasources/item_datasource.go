// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package datasources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/planeopscc/terraform-provider-protonpass/internal/passcli"
)

var _ datasource.DataSource = &ItemDataSource{}

type ItemDataSource struct {
	client *passcli.Client
}

type ItemDataSourceModel struct {
	ItemID             types.String   `tfsdk:"item_id"`
	ShareID            types.String   `tfsdk:"share_id"`
	Name               types.String   `tfsdk:"name"`
	Type               types.String   `tfsdk:"type"`
	Title              types.String   `tfsdk:"title"`
	Note               types.String   `tfsdk:"note"`
	CreateTime         types.String   `tfsdk:"create_time"`
	ModifyTime         types.String   `tfsdk:"modify_time"`
	Username           types.String   `tfsdk:"username"`
	Email              types.String   `tfsdk:"email"`
	Password           types.String   `tfsdk:"password"`
	URLs               []types.String `tfsdk:"urls"`
	TOTPUri            types.String   `tfsdk:"totp_uri"`
	CardholderName     types.String   `tfsdk:"cardholder_name"`
	Number             types.String   `tfsdk:"number"`
	VerificationNumber types.String   `tfsdk:"verification_number"`
	ExpirationDate     types.String   `tfsdk:"expiration_date"`
	PIN                types.String   `tfsdk:"pin"`
	SSID               types.String   `tfsdk:"ssid"`
	Security           types.String   `tfsdk:"security"`
	PrivateKey         types.String   `tfsdk:"private_key"`
	PublicKey          types.String   `tfsdk:"public_key"`
	// Identity fields omitted for brevity here unless necessary, let's include a few common ones
	FullName    types.String `tfsdk:"full_name"`
	PhoneNumber types.String `tfsdk:"phone_number"`
}

func NewItemDataSource() datasource.DataSource {
	return &ItemDataSource{}
}

func (d *ItemDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_item"
}

func (d *ItemDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a specific Proton Pass item by name and share ID.",
		Attributes: map[string]schema.Attribute{
			"item_id": schema.StringAttribute{
				MarkdownDescription: "Item ID.",
				Computed:            true,
			},
			"share_id": schema.StringAttribute{
				MarkdownDescription: "Share ID of the vault.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name (title) of the item to lookup.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Filter by item type (login, note, credit-card...)",
				Optional:            true,
				Computed:            true,
			},
			"title":               schema.StringAttribute{Computed: true, MarkdownDescription: "Title of the item."},
			"note":                schema.StringAttribute{Computed: true, MarkdownDescription: "Note content."},
			"create_time":         schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp."},
			"modify_time":         schema.StringAttribute{Computed: true, MarkdownDescription: "Last modification timestamp."},
			"username":            schema.StringAttribute{Computed: true, MarkdownDescription: "Username."},
			"email":               schema.StringAttribute{Computed: true, MarkdownDescription: "Email."},
			"password":            schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Password."},
			"urls":                schema.ListAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "URLs."},
			"totp_uri":            schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "TOTP URI."},
			"cardholder_name":     schema.StringAttribute{Computed: true, MarkdownDescription: "Cardholder name."},
			"number":              schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Card number."},
			"verification_number": schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Verification number (CVV)."},
			"expiration_date":     schema.StringAttribute{Computed: true, MarkdownDescription: "Expiration date."},
			"pin":                 schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "PIN."},
			"ssid":                schema.StringAttribute{Computed: true, MarkdownDescription: "WiFi SSID."},
			"security":            schema.StringAttribute{Computed: true, MarkdownDescription: "WiFi Security."},
			"private_key":         schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "SSH Private Key."},
			"public_key":          schema.StringAttribute{Computed: true, MarkdownDescription: "SSH Public Key."},
			"full_name":           schema.StringAttribute{Computed: true, MarkdownDescription: "Identity Full Name."},
			"phone_number":        schema.StringAttribute{Computed: true, MarkdownDescription: "Identity Phone Number."},
		},
	}
}

func (d *ItemDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ItemDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ItemDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	shareID := data.ShareID.ValueString()
	name := data.Name.ValueString()
	searchType := data.Type.ValueString()

	items, err := d.client.ListItemsInVault(ctx, shareID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list items for lookup", err.Error())
		return
	}

	var foundItem *passcli.ItemJSON
	for _, v := range items {
		if strings.EqualFold(v.Title, name) {
			if searchType == "" || strings.EqualFold(v.Type, searchType) {
				vCopy := v
				foundItem = &vCopy
				break
			}
		}
	}

	if foundItem == nil {
		resp.Diagnostics.AddError("Item Not Found", fmt.Sprintf("No item found with name %q in share %q", name, shareID))
		return
	}

	// Now fetch the full item using GetItem
	rawItem, err := d.client.GetItem(ctx, foundItem.ItemID, "", shareID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get item details", err.Error())
		return
	}

	item := passcli.FlattenItem(*rawItem)

	data.ItemID = types.StringValue(item.ItemID)
	data.Title = types.StringValue(item.Title)
	data.Type = types.StringValue(item.Type)
	data.Note = types.StringValue(item.Note)
	data.CreateTime = types.StringValue(item.CreateTime)
	data.ModifyTime = types.StringValue(item.ModifyTime)

	// String fields helpers
	setString := func(v string) types.String {
		if v == "" {
			return types.StringNull()
		}
		return types.StringValue(v)
	}

	data.Username = setString(item.Username)
	data.Email = setString(item.Email)
	data.Password = setString(item.Password)
	data.TOTPUri = setString(item.TOTPUri)

	data.CardholderName = setString(item.CardholderName)
	data.Number = setString(item.Number)
	data.VerificationNumber = setString(item.VerificationNumber)
	data.ExpirationDate = setString(item.ExpirationDate)
	data.PIN = setString(item.PIN)

	data.SSID = setString(item.SSID)
	data.Security = setString(item.Security)

	data.PrivateKey = setString(item.PrivateKey)
	data.PublicKey = setString(item.PublicKey)

	data.FullName = setString(item.FullName)
	data.PhoneNumber = setString(item.PhoneNumber)

	if len(item.URLs) > 0 {
		var urlVals []types.String
		for _, u := range item.URLs {
			urlVals = append(urlVals, types.StringValue(u))
		}
		data.URLs = urlVals
	} else {
		data.URLs = nil
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
