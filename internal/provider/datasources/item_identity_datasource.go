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

var _ datasource.DataSource = &ItemIdentityDataSource{}

type ItemIdentityDataSource struct {
	client *passcli.Client
}

type ItemIdentityDataSourceModel struct {
	ItemID          types.String `tfsdk:"item_id"`
	ShareID         types.String `tfsdk:"share_id"`
	Title           types.String `tfsdk:"title"`
	FullName        types.String `tfsdk:"full_name"`
	Email           types.String `tfsdk:"email"`
	Company         types.String `tfsdk:"company"`
	PhoneNumber     types.String `tfsdk:"phone_number"`
	FirstName       types.String `tfsdk:"first_name"`
	MiddleName      types.String `tfsdk:"middle_name"`
	LastName        types.String `tfsdk:"last_name"`
	Birthdate       types.String `tfsdk:"birthdate"`
	Gender          types.String `tfsdk:"gender"`
	Organization    types.String `tfsdk:"organization"`
	StreetAddress   types.String `tfsdk:"street_address"`
	ZipOrPostalCode types.String `tfsdk:"zip_or_postal_code"`
	City            types.String `tfsdk:"city"`
	CountryOrRegion types.String `tfsdk:"country"`
	SocialSecurity  types.String `tfsdk:"social_security_number"`
	PassportNumber  types.String `tfsdk:"passport_number"`
	LicenseNumber   types.String `tfsdk:"license_number"`
	Website         types.String `tfsdk:"website"`
	JobTitle        types.String `tfsdk:"job_title"`
	WorkEmail       types.String `tfsdk:"work_email"`
	WorkPhoneNumber types.String `tfsdk:"work_phone_number"`
	CreateTime      types.String `tfsdk:"create_time"`
	ModifyTime      types.String `tfsdk:"modify_time"`
}

func NewItemIdentityDataSource() datasource.DataSource {
	return &ItemIdentityDataSource{}
}

func (d *ItemIdentityDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_item_identity"
}

func (d *ItemIdentityDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a specific Proton Pass Identity item.",
		Attributes: map[string]schema.Attribute{
			"item_id":                schema.StringAttribute{Required: true, MarkdownDescription: "Item ID."},
			"share_id":               schema.StringAttribute{Required: true, MarkdownDescription: "Share ID of the vault."},
			"title":                  schema.StringAttribute{Computed: true, MarkdownDescription: "Title."},
			"full_name":              schema.StringAttribute{Computed: true, MarkdownDescription: "Full Name."},
			"email":                  schema.StringAttribute{Computed: true, MarkdownDescription: "Email."},
			"company":                schema.StringAttribute{Computed: true, MarkdownDescription: "Company."},
			"phone_number":           schema.StringAttribute{Computed: true, MarkdownDescription: "Phone number."},
			"first_name":             schema.StringAttribute{Computed: true, MarkdownDescription: "First name."},
			"middle_name":            schema.StringAttribute{Computed: true, MarkdownDescription: "Middle name."},
			"last_name":              schema.StringAttribute{Computed: true, MarkdownDescription: "Last name."},
			"birthdate":              schema.StringAttribute{Computed: true, MarkdownDescription: "Birth date."},
			"gender":                 schema.StringAttribute{Computed: true, MarkdownDescription: "Gender."},
			"organization":           schema.StringAttribute{Computed: true, MarkdownDescription: "Organization."},
			"street_address":         schema.StringAttribute{Computed: true, MarkdownDescription: "Street address."},
			"zip_or_postal_code":     schema.StringAttribute{Computed: true, MarkdownDescription: "ZIP/Postal code."},
			"city":                   schema.StringAttribute{Computed: true, MarkdownDescription: "City."},
			"country":                schema.StringAttribute{Computed: true, MarkdownDescription: "Country."},
			"social_security_number": schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "SSN."},
			"passport_number":        schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Passport."},
			"license_number":         schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "License number."},
			"website":                schema.StringAttribute{Computed: true, MarkdownDescription: "Website."},
			"job_title":              schema.StringAttribute{Computed: true, MarkdownDescription: "Job title."},
			"work_email":             schema.StringAttribute{Computed: true, MarkdownDescription: "Work email."},
			"work_phone_number":      schema.StringAttribute{Computed: true, MarkdownDescription: "Work phone."},
			"create_time":            schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp."},
			"modify_time":            schema.StringAttribute{Computed: true, MarkdownDescription: "Last modification timestamp."},
		},
	}
}

func (d *ItemIdentityDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ItemIdentityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ItemIdentityDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	item, err := d.client.GetItem(ctx, data.ItemID.ValueString(), "", data.ShareID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Identity item", err.Error())
		return
	}

	data.Title = types.StringValue(item.Content.Title)
	data.CreateTime = types.StringValue(item.CreateTime)
	data.ModifyTime = types.StringValue(item.ModifyTime)

	if item.Content.Content.Identity != nil {
		id := item.Content.Content.Identity
		data.FullName = types.StringValue(id.FullName)
		data.Email = types.StringValue(id.Email)
		data.Company = types.StringValue(id.Company)
		data.PhoneNumber = types.StringValue(id.PhoneNumber)
		data.FirstName = types.StringValue(id.FirstName)
		data.MiddleName = types.StringValue(id.MiddleName)
		data.LastName = types.StringValue(id.LastName)
		data.Birthdate = types.StringValue(id.Birthdate)
		data.Gender = types.StringValue(id.Gender)
		data.Organization = types.StringValue(id.Organization)
		data.StreetAddress = types.StringValue(id.StreetAddress)
		data.ZipOrPostalCode = types.StringValue(id.ZipOrPostalCode)
		data.City = types.StringValue(id.City)
		data.CountryOrRegion = types.StringValue(id.CountryOrRegion)
		data.SocialSecurity = types.StringValue(id.SocialSecurity)
		data.PassportNumber = types.StringValue(id.PassportNumber)
		data.LicenseNumber = types.StringValue(id.LicenseNumber)
		data.Website = types.StringValue(id.Website)
		data.JobTitle = types.StringValue(id.JobTitle)
		data.WorkEmail = types.StringValue(id.WorkEmail)
		data.WorkPhoneNumber = types.StringValue(id.WorkPhoneNumber)
	} else {
		resp.Diagnostics.AddError("Invalid Item Type", "The specified item is not an Identity item.")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
