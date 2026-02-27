// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"
	"encoding/json"
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

var _ resource.Resource = &ItemIdentityResource{}
var _ resource.ResourceWithImportState = &ItemIdentityResource{}

type ItemIdentityResource struct {
	client *passcli.Client
}

type ItemIdentityResourceModel struct {
	ItemID              types.String `tfsdk:"item_id"`
	ShareID             types.String `tfsdk:"share_id"`
	Title               types.String `tfsdk:"title"`
	FullName            types.String `tfsdk:"full_name"`
	Email               types.String `tfsdk:"email"`
	PhoneNumber         types.String `tfsdk:"phone_number"`
	FirstName           types.String `tfsdk:"first_name"`
	MiddleName          types.String `tfsdk:"middle_name"`
	LastName            types.String `tfsdk:"last_name"`
	Birthdate           types.String `tfsdk:"birthdate"`
	Gender              types.String `tfsdk:"gender"`
	Organization        types.String `tfsdk:"organization"`
	StreetAddress       types.String `tfsdk:"street_address"`
	ZipOrPostalCode     types.String `tfsdk:"zip_or_postal_code"`
	City                types.String `tfsdk:"city"`
	StateOrProvince     types.String `tfsdk:"state_or_province"`
	CountryOrRegion     types.String `tfsdk:"country_or_region"`
	SsnWO               types.String `tfsdk:"ssn_wo"`
	SsnWOVersion        types.Int64  `tfsdk:"ssn_wo_version"`
	PassportNumberWO    types.String `tfsdk:"passport_number_wo"`
	PassportNumberWOVer types.Int64  `tfsdk:"passport_number_wo_version"`
	LicenseNumberWO     types.String `tfsdk:"license_number_wo"`
	LicenseNumberWOVer  types.Int64  `tfsdk:"license_number_wo_version"`
	Website             types.String `tfsdk:"website"`
	Company             types.String `tfsdk:"company"`
	JobTitle            types.String `tfsdk:"job_title"`
	WorkEmail           types.String `tfsdk:"work_email"`
	WorkPhoneNumber     types.String `tfsdk:"work_phone_number"`
	CreateTime          types.String `tfsdk:"create_time"`
	ModifyTime          types.String `tfsdk:"modify_time"`
}

func NewItemIdentityResource() resource.Resource {
	return &ItemIdentityResource{}
}

func (r *ItemIdentityResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_item_identity"
}

func (r *ItemIdentityResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Proton Pass identity item.",
		Attributes: map[string]schema.Attribute{
			"item_id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the item.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"share_id": schema.StringAttribute{
				MarkdownDescription: "Share ID of the vault.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Title of the identity item.",
				Required:            true,
			},
			"full_name":          schema.StringAttribute{Optional: true, MarkdownDescription: "Full name."},
			"email":              schema.StringAttribute{Optional: true, MarkdownDescription: "Email address."},
			"phone_number":       schema.StringAttribute{Optional: true, MarkdownDescription: "Phone number."},
			"first_name":         schema.StringAttribute{Optional: true, MarkdownDescription: "First name."},
			"middle_name":        schema.StringAttribute{Optional: true, MarkdownDescription: "Middle name."},
			"last_name":          schema.StringAttribute{Optional: true, MarkdownDescription: "Last name."},
			"birthdate":          schema.StringAttribute{Optional: true, MarkdownDescription: "Birth date."},
			"gender":             schema.StringAttribute{Optional: true, MarkdownDescription: "Gender."},
			"organization":       schema.StringAttribute{Optional: true, MarkdownDescription: "Organization."},
			"street_address":     schema.StringAttribute{Optional: true, MarkdownDescription: "Street address."},
			"zip_or_postal_code": schema.StringAttribute{Optional: true, MarkdownDescription: "ZIP/Postal code."},
			"city":               schema.StringAttribute{Optional: true, MarkdownDescription: "City."},
			"state_or_province":  schema.StringAttribute{Optional: true, MarkdownDescription: "State or province."},
			"country_or_region":  schema.StringAttribute{Optional: true, MarkdownDescription: "Country or region."},
			"website":            schema.StringAttribute{Optional: true, MarkdownDescription: "Website."},
			"company":            schema.StringAttribute{Optional: true, MarkdownDescription: "Company."},
			"job_title":          schema.StringAttribute{Optional: true, MarkdownDescription: "Job title."},
			"work_email":         schema.StringAttribute{Optional: true, MarkdownDescription: "Work email."},
			"work_phone_number":  schema.StringAttribute{Optional: true, MarkdownDescription: "Work phone number."},
			"ssn_wo": schema.StringAttribute{
				MarkdownDescription: "Social Security Number (write-only).",
				Optional:            true, WriteOnly: true,
			},
			"ssn_wo_version": schema.Int64Attribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"passport_number_wo": schema.StringAttribute{
				MarkdownDescription: "Passport number (write-only).",
				Optional:            true, WriteOnly: true,
			},
			"passport_number_wo_version": schema.Int64Attribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"license_number_wo": schema.StringAttribute{
				MarkdownDescription: "License number (write-only).",
				Optional:            true, WriteOnly: true,
			},
			"license_number_wo_version": schema.Int64Attribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"create_time": schema.StringAttribute{
				Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"modify_time": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *ItemIdentityResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// buildIdentityTemplate builds the JSON template from resource data.
func (r *ItemIdentityResource) buildIdentityTemplate(data ItemIdentityResourceModel) map[string]string {
	tmpl := map[string]string{
		"title": data.Title.ValueString(),
	}
	setIf := func(k string, v types.String) {
		if !v.IsNull() && !v.IsUnknown() {
			tmpl[k] = v.ValueString()
		}
	}
	setIf("full_name", data.FullName)
	setIf("email", data.Email)
	setIf("phone_number", data.PhoneNumber)
	setIf("first_name", data.FirstName)
	setIf("middle_name", data.MiddleName)
	setIf("last_name", data.LastName)
	setIf("birthdate", data.Birthdate)
	setIf("gender", data.Gender)
	setIf("organization", data.Organization)
	setIf("street_address", data.StreetAddress)
	setIf("zip_or_postal_code", data.ZipOrPostalCode)
	setIf("city", data.City)
	setIf("state_or_province", data.StateOrProvince)
	setIf("country_or_region", data.CountryOrRegion)
	setIf("website", data.Website)
	setIf("company", data.Company)
	setIf("job_title", data.JobTitle)
	setIf("work_email", data.WorkEmail)
	setIf("work_phone_number", data.WorkPhoneNumber)
	return tmpl
}

func (r *ItemIdentityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ItemIdentityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl := r.buildIdentityTemplate(data)

	// Read write-only fields from config.
	readWO := func(attr string) string {
		var v types.String
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(attr), &v)...)
		if !v.IsNull() && !v.IsUnknown() {
			return v.ValueString()
		}
		return ""
	}
	if ssn := readWO("ssn_wo"); ssn != "" {
		tmpl["social_security_number"] = ssn
	}
	if pn := readWO("passport_number_wo"); pn != "" {
		tmpl["passport_number"] = pn
	}
	if ln := readWO("license_number_wo"); ln != "" {
		tmpl["license_number"] = ln
	}

	templateJSON, err := json.Marshal(tmpl)
	if err != nil {
		resp.Diagnostics.AddError("Failed to build identity template", err.Error())
		return
	}

	tflog.Debug(ctx, "creating item_identity", map[string]interface{}{
		"share_id": data.ShareID.ValueString(),
		"title":    data.Title.ValueString(),
	})

	item, err := r.client.CreateItemIdentity(ctx, data.ShareID.ValueString(), templateJSON)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create identity item", err.Error())
		return
	}

	data.ItemID = types.StringValue(item.ItemID)
	data.ShareID = types.StringValue(item.ShareID)
	data.CreateTime = types.StringValue(item.CreateTime)
	data.ModifyTime = types.StringValue(item.ModifyTime)
	if data.SsnWOVersion.IsNull() || data.SsnWOVersion.IsUnknown() {
		data.SsnWOVersion = types.Int64Value(0)
	}
	if data.PassportNumberWOVer.IsNull() || data.PassportNumberWOVer.IsUnknown() {
		data.PassportNumberWOVer = types.Int64Value(0)
	}
	if data.LicenseNumberWOVer.IsNull() || data.LicenseNumberWOVer.IsUnknown() {
		data.LicenseNumberWOVer = types.Int64Value(0)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ItemIdentityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ItemIdentityResourceModel
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
		resp.Diagnostics.AddError("Failed to read identity item", err.Error())
		return
	}

	data.Title = types.StringValue(item.Title)
	data.CreateTime = types.StringValue(item.CreateTime)
	data.ModifyTime = types.StringValue(item.ModifyTime)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ItemIdentityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ItemIdentityResourceModel
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
		err := r.client.UpdateItemLogin(ctx, state.ItemID.ValueString(), state.ShareID.ValueString(), fields)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update identity item", err.Error())
			return
		}
	}

	item, err := r.client.ReadItemLogin(ctx, state.ItemID.ValueString(), state.ShareID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read item after update", err.Error())
		return
	}

	plan.ItemID = state.ItemID
	plan.ShareID = state.ShareID
	plan.Title = types.StringValue(item.Title)
	plan.CreateTime = types.StringValue(item.CreateTime)
	plan.ModifyTime = types.StringValue(item.ModifyTime)
	if plan.SsnWOVersion.IsNull() || plan.SsnWOVersion.IsUnknown() {
		plan.SsnWOVersion = state.SsnWOVersion
	}
	if plan.PassportNumberWOVer.IsNull() || plan.PassportNumberWOVer.IsUnknown() {
		plan.PassportNumberWOVer = state.PassportNumberWOVer
	}
	if plan.LicenseNumberWOVer.IsNull() || plan.LicenseNumberWOVer.IsUnknown() {
		plan.LicenseNumberWOVer = state.LicenseNumberWOVer
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ItemIdentityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ItemIdentityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.DeleteItemLogin(ctx, data.ItemID.ValueString(), data.ShareID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete identity item", err.Error())
		return
	}
}

func (r *ItemIdentityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
