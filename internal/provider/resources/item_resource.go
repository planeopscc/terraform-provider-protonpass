// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/planeopscc/terraform-provider-protonpass/internal/passcli"
)

var _ resource.Resource = &ItemResource{}
var _ resource.ResourceWithImportState = &ItemResource{}

// ItemResource manages a Proton Pass item (login, note, credit-card, wifi, ssh-key, identity).
type ItemResource struct {
	client *passcli.Client
}

// ItemResourceModel describes the Terraform state for a Proton Pass item.
type ItemResourceModel struct {
	ItemID             types.String `tfsdk:"item_id"`
	ShareID            types.String `tfsdk:"share_id"`
	Type               types.String `tfsdk:"type"`
	Title              types.String `tfsdk:"title"`
	DestroyPermanently types.Bool   `tfsdk:"destroy_permanently"`
	CreateTime         types.String `tfsdk:"create_time"`
	ModifyTime         types.String `tfsdk:"modify_time"`

	// Login
	Username          types.String `tfsdk:"username"`
	Email             types.String `tfsdk:"email"`
	Password          types.String `tfsdk:"password"`
	PasswordWO        types.String `tfsdk:"password_wo"`
	PasswordWOVersion types.Int64  `tfsdk:"password_wo_version"`
	URLs              types.List   `tfsdk:"urls"`

	// Note
	Note          types.String `tfsdk:"note"`
	NoteWO        types.String `tfsdk:"note_wo"`
	NoteWOVersion types.Int64  `tfsdk:"note_wo_version"`

	// Credit Card
	CardholderName     types.String `tfsdk:"cardholder_name"`
	Number             types.String `tfsdk:"number"`
	VerificationNumber types.String `tfsdk:"verification_number"`
	ExpirationDate     types.String `tfsdk:"expiration_date"`
	PIN                types.String `tfsdk:"pin"`

	// WiFi
	SSID     types.String `tfsdk:"ssid"`
	Security types.String `tfsdk:"security"`

	// SSH Key
	PrivateKey types.String `tfsdk:"private_key"`
	PublicKey  types.String `tfsdk:"public_key"`
	KeyType    types.String `tfsdk:"key_type"`
	Comment    types.String `tfsdk:"comment"`
	Generate   types.Bool   `tfsdk:"generate"`

	// Identity
	FullName                types.String `tfsdk:"full_name"`
	PhoneNumber             types.String `tfsdk:"phone_number"`
	FirstName               types.String `tfsdk:"first_name"`
	MiddleName              types.String `tfsdk:"middle_name"`
	LastName                types.String `tfsdk:"last_name"`
	Birthdate               types.String `tfsdk:"birthdate"`
	Gender                  types.String `tfsdk:"gender"`
	Organization            types.String `tfsdk:"organization"`
	StreetAddress           types.String `tfsdk:"street_address"`
	ZipOrPostalCode         types.String `tfsdk:"zip_or_postal_code"`
	City                    types.String `tfsdk:"city"`
	StateOrProvince         types.String `tfsdk:"state_or_province"`
	CountryOrRegion         types.String `tfsdk:"country_or_region"`
	Ssn                     types.String `tfsdk:"ssn"`
	SsnWO                   types.String `tfsdk:"ssn_wo"`
	SsnWOVersion            types.Int64  `tfsdk:"ssn_wo_version"`
	PassportNumber          types.String `tfsdk:"passport_number"`
	PassportNumberWO        types.String `tfsdk:"passport_number_wo"`
	PassportNumberWOVersion types.Int64  `tfsdk:"passport_number_wo_version"`
	LicenseNumber           types.String `tfsdk:"license_number"`
	LicenseNumberWO         types.String `tfsdk:"license_number_wo"`
	LicenseNumberWOVersion  types.Int64  `tfsdk:"license_number_wo_version"`
	Website                 types.String `tfsdk:"website"`
	Company                 types.String `tfsdk:"company"`
	JobTitle                types.String `tfsdk:"job_title"`
	WorkEmail               types.String `tfsdk:"work_email"`
	WorkPhoneNumber         types.String `tfsdk:"work_phone_number"`
}

// NewItemResource returns a new ItemResource factory.
func NewItemResource() resource.Resource {
	return &ItemResource{}
}

func (r *ItemResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_item"
}

func (r *ItemResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a generic Proton Pass item (e.g. login, note, identity, etc).",
		Attributes: map[string]schema.Attribute{
			"item_id":             schema.StringAttribute{MarkdownDescription: "The unique identifier of the item.", Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"share_id":            schema.StringAttribute{MarkdownDescription: "The share ID of the vault containing this item.", Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"type":                schema.StringAttribute{MarkdownDescription: "The item type: `login`, `note`, `credit-card`, `wifi`, `ssh-key`, or `identity`.", Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"title":               schema.StringAttribute{MarkdownDescription: "The title of the item.", Required: true},
			"destroy_permanently": schema.BoolAttribute{MarkdownDescription: "If true, the item is permanently deleted on destroy instead of being moved to trash.", Optional: true},
			"create_time":         schema.StringAttribute{MarkdownDescription: "Timestamp when the item was created.", Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"modify_time":         schema.StringAttribute{MarkdownDescription: "Timestamp when the item was last modified.", Computed: true},

			// Fields
			"username":            schema.StringAttribute{MarkdownDescription: "Username for login items.", Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"email":               schema.StringAttribute{MarkdownDescription: "Email address for login or identity items.", Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"password":            schema.StringAttribute{MarkdownDescription: "Password for login or WiFi items.", Optional: true, Computed: true, Sensitive: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"password_wo":         schema.StringAttribute{MarkdownDescription: "Write-only password. Use this instead of `password` to avoid storing the password in state.", Optional: true, Sensitive: true, WriteOnly: true},
			"password_wo_version": schema.Int64Attribute{MarkdownDescription: "Version counter for `password_wo`. Increment to trigger a password update.", Optional: true},
			"urls":                schema.ListAttribute{MarkdownDescription: "List of URLs associated with the login item.", Optional: true, ElementType: types.StringType},

			"note":            schema.StringAttribute{MarkdownDescription: "Note content.", Optional: true, Computed: true, Sensitive: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"note_wo":         schema.StringAttribute{MarkdownDescription: "Write-only note content.", Optional: true, Sensitive: true, WriteOnly: true},
			"note_wo_version": schema.Int64Attribute{MarkdownDescription: "Version counter for `note_wo`. Increment to trigger a note update.", Optional: true},

			"cardholder_name":     schema.StringAttribute{MarkdownDescription: "Cardholder name for credit card items.", Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"number":              schema.StringAttribute{MarkdownDescription: "Card number for credit card items.", Optional: true, Computed: true, Sensitive: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"verification_number": schema.StringAttribute{MarkdownDescription: "CVV/verification number for credit card items.", Optional: true, Computed: true, Sensitive: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"expiration_date":     schema.StringAttribute{MarkdownDescription: "Expiration date for credit card items (format: `YYYY-MM`).", Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"pin":                 schema.StringAttribute{MarkdownDescription: "PIN for credit card items.", Optional: true, Computed: true, Sensitive: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},

			"ssid":     schema.StringAttribute{MarkdownDescription: "SSID for WiFi items.", Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"security": schema.StringAttribute{MarkdownDescription: "Security type for WiFi items (e.g. `WPA2`, `WPA3`).", Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},

			"private_key": schema.StringAttribute{MarkdownDescription: "Private key for SSH key items.", Optional: true, Computed: true, Sensitive: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"public_key":  schema.StringAttribute{MarkdownDescription: "Public key for SSH key items.", Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"key_type":    schema.StringAttribute{MarkdownDescription: "Key type for SSH key generation (e.g. `ed25519`).", Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"comment":     schema.StringAttribute{MarkdownDescription: "Comment for SSH key generation.", Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"generate":    schema.BoolAttribute{MarkdownDescription: "If true, generates a new SSH key pair.", Optional: true},

			"full_name":                  schema.StringAttribute{MarkdownDescription: "Full name for identity items.", Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"phone_number":               schema.StringAttribute{MarkdownDescription: "Phone number for identity items.", Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"first_name":                 schema.StringAttribute{MarkdownDescription: "First name for identity items.", Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"middle_name":                schema.StringAttribute{MarkdownDescription: "Middle name for identity items.", Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"last_name":                  schema.StringAttribute{MarkdownDescription: "Last name for identity items.", Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"birthdate":                  schema.StringAttribute{MarkdownDescription: "Birthdate for identity items.", Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"gender":                     schema.StringAttribute{MarkdownDescription: "Gender for identity items.", Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"organization":               schema.StringAttribute{MarkdownDescription: "Organization for identity items.", Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"street_address":             schema.StringAttribute{MarkdownDescription: "Street address for identity items.", Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"zip_or_postal_code":         schema.StringAttribute{MarkdownDescription: "ZIP or postal code for identity items.", Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"city":                       schema.StringAttribute{MarkdownDescription: "City for identity items.", Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"state_or_province":          schema.StringAttribute{MarkdownDescription: "State or province for identity items.", Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"country_or_region":          schema.StringAttribute{MarkdownDescription: "Country or region for identity items.", Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"ssn":                        schema.StringAttribute{MarkdownDescription: "Social Security Number for identity items.", Optional: true, Computed: true, Sensitive: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"ssn_wo":                     schema.StringAttribute{MarkdownDescription: "Write-only SSN. Use this instead of `ssn` to avoid storing it in state.", Optional: true, Sensitive: true, WriteOnly: true},
			"ssn_wo_version":             schema.Int64Attribute{MarkdownDescription: "Version counter for `ssn_wo`.", Optional: true},
			"passport_number":            schema.StringAttribute{MarkdownDescription: "Passport number for identity items.", Optional: true, Computed: true, Sensitive: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"passport_number_wo":         schema.StringAttribute{MarkdownDescription: "Write-only passport number.", Optional: true, Sensitive: true, WriteOnly: true},
			"passport_number_wo_version": schema.Int64Attribute{MarkdownDescription: "Version counter for `passport_number_wo`.", Optional: true},
			"license_number":             schema.StringAttribute{MarkdownDescription: "License number for identity items.", Optional: true, Computed: true, Sensitive: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"license_number_wo":          schema.StringAttribute{MarkdownDescription: "Write-only license number.", Optional: true, Sensitive: true, WriteOnly: true},
			"license_number_wo_version":  schema.Int64Attribute{MarkdownDescription: "Version counter for `license_number_wo`.", Optional: true},
			"website":                    schema.StringAttribute{MarkdownDescription: "Website URL for identity items.", Optional: true},
			"company":                    schema.StringAttribute{MarkdownDescription: "Company name for identity items.", Optional: true},
			"job_title":                  schema.StringAttribute{MarkdownDescription: "Job title for identity items.", Optional: true},
			"work_email":                 schema.StringAttribute{MarkdownDescription: "Work email for identity items.", Optional: true},
			"work_phone_number":          schema.StringAttribute{MarkdownDescription: "Work phone number for identity items.", Optional: true},
		},
	}
}

func (r *ItemResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func getCreatedSecret(ctx context.Context, req resource.CreateRequest, planVal types.String, woAttr string) string {
	var v types.String
	req.Config.GetAttribute(ctx, path.Root(woAttr), &v)
	if !v.IsNull() && !v.IsUnknown() && v.ValueString() != "" {
		return v.ValueString()
	}
	if !planVal.IsNull() && !planVal.IsUnknown() {
		return planVal.ValueString()
	}
	return ""
}

func getURLs(ctx context.Context, data ItemResourceModel) []string {
	if data.URLs.IsNull() || data.URLs.IsUnknown() {
		return nil
	}
	var urls []string
	data.URLs.ElementsAs(ctx, &urls, false)
	return urls
}

func (r *ItemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ItemResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	shareID := data.ShareID.ValueString()
	title := data.Title.ValueString()
	itemType := data.Type.ValueString()

	var item *passcli.ItemJSON
	var err error

	// 1. Check if item exists in trash
	trashedItems, _ := r.client.ListTrashedItems(ctx, shareID)
	var existingID string
	for _, ti := range trashedItems {
		if ti.Title == title && ti.Type == itemType {
			existingID = ti.ItemID
			break
		}
	}

	if existingID != "" {
		err = r.client.RestoreItem(ctx, existingID, shareID)
		if err != nil {
			resp.Diagnostics.AddError("Failed to restore item from trash", err.Error())
			return
		}
		// Fetch the restored item
		item, err = r.client.ReadItem(ctx, existingID, shareID)
		if err != nil {
			resp.Diagnostics.AddError("Failed to read restored item", err.Error())
			return
		}
	} else {
		// 2. Normal Creation
		switch itemType {
		case "login":
			pw := getCreatedSecret(ctx, req, data.Password, "password_wo")
			urls := getURLs(ctx, data)
			item, err = r.client.CreateItemLogin(ctx, shareID, title, data.Username.ValueString(), pw, data.Email.ValueString(), urls)
		case "note":
			note := getCreatedSecret(ctx, req, data.Note, "note_wo")
			item, err = r.client.CreateItemNote(ctx, shareID, title, note)
		case "credit-card":
			item, err = r.client.CreateItemCreditCard(ctx, shareID, title, data.CardholderName.ValueString(), data.Number.ValueString(), data.VerificationNumber.ValueString(), data.ExpirationDate.ValueString(), data.PIN.ValueString())
		case "wifi":
			item, err = r.client.CreateItemWiFi(ctx, shareID, title, data.SSID.ValueString(), getCreatedSecret(ctx, req, data.Password, "password_wo"), data.Security.ValueString())
		case "ssh-key":
			isGen := data.Generate.ValueBool()
			if isGen {
				item, err = r.client.CreateItemSSHKey(ctx, shareID, title, data.KeyType.ValueString(), data.Comment.ValueString())
			} else {
				// Import from provided private key
				pk := data.PrivateKey.ValueString()
				tmpFile, ferr := passcli.WriteTempFile("protonpass-ssh-import-*.key", []byte(pk))
				if ferr != nil {
					resp.Diagnostics.AddError("Failed to write temp SSH key", ferr.Error())
					return
				}
				defer os.Remove(tmpFile)
				item, err = r.client.CreateItemSSHKeyImport(ctx, shareID, title, tmpFile)
			}
		case "identity":
			tmpl := map[string]string{"title": title}
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

			if ssn := getCreatedSecret(ctx, req, data.Ssn, "ssn_wo"); ssn != "" {
				tmpl["social_security_number"] = ssn
			}
			if pn := getCreatedSecret(ctx, req, data.PassportNumber, "passport_number_wo"); pn != "" {
				tmpl["passport_number"] = pn
			}
			if ln := getCreatedSecret(ctx, req, data.LicenseNumber, "license_number_wo"); ln != "" {
				tmpl["license_number"] = ln
			}
			templateJSON, _ := json.Marshal(tmpl)
			item, err = r.client.CreateItemIdentity(ctx, shareID, templateJSON)
		default:
			resp.Diagnostics.AddError("Unsupported item type", itemType)
			return
		}
	}

	if err != nil {
		resp.Diagnostics.AddError("Failed to process item", err.Error())
		return
	}

	if item == nil {
		resp.Diagnostics.AddError("Failed to process item", "Item object is nil after creation/restoration")
		return
	}

	// 3. Update with plan values (important if restored or if creation didn't set everything)
	// We reuse the update logic here or just rely on the next Read.
	// To be safe and ensure all fields match config, let's trigger a field update.
	fields := make(map[string]string)
	switch itemType {
	case "login":
		fields["username"] = data.Username.ValueString()
		fields["email"] = data.Email.ValueString()
	case "wifi":
		fields["ssid"] = data.SSID.ValueString()
		fields["security"] = data.Security.ValueString()
	}
	// ... add other fields as needed for a full sync ...
	if len(fields) > 0 {
		_ = r.client.UpdateItem(ctx, item.ItemID, shareID, fields)
	}

	resp.Diagnostics.Append(r.mapItemToModel(ctx, item, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// setString converts an empty string to types.StringNull, otherwise types.StringValue.
func setString(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func (r *ItemResource) mapItemToModel(ctx context.Context, item *passcli.ItemJSON, data *ItemResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// Initialize all computed/optional fields to null to avoid "unknown value" errors after apply.
	// Terraform requires all attributes defined in the schema to have a known value in the state.
	data.Username = types.StringNull()
	data.Email = types.StringNull()
	data.Password = types.StringNull()
	data.URLs = types.ListNull(types.StringType)
	data.Note = types.StringNull()
	data.CardholderName = types.StringNull()
	data.Number = types.StringNull()
	data.VerificationNumber = types.StringNull()
	data.ExpirationDate = types.StringNull()
	data.PIN = types.StringNull()
	data.SSID = types.StringNull()
	data.Security = types.StringNull()
	data.PrivateKey = types.StringNull()
	data.PublicKey = types.StringNull()
	data.FullName = types.StringNull()
	data.PhoneNumber = types.StringNull()
	data.FirstName = types.StringNull()
	data.MiddleName = types.StringNull()
	data.LastName = types.StringNull()
	data.Birthdate = types.StringNull()
	data.Gender = types.StringNull()
	data.Organization = types.StringNull()
	data.StreetAddress = types.StringNull()
	data.ZipOrPostalCode = types.StringNull()
	data.City = types.StringNull()
	data.StateOrProvince = types.StringNull()
	data.CountryOrRegion = types.StringNull()
	data.Ssn = types.StringNull()
	data.PassportNumber = types.StringNull()
	data.LicenseNumber = types.StringNull()
	data.Website = types.StringNull()
	data.Company = types.StringNull()
	data.JobTitle = types.StringNull()
	data.WorkEmail = types.StringNull()
	data.WorkPhoneNumber = types.StringNull()

	data.Title = types.StringValue(item.Title)
	data.Note = setString(item.Note)
	data.CreateTime = types.StringValue(item.CreateTime)
	data.ModifyTime = types.StringValue(item.ModifyTime)
	data.ItemID = types.StringValue(item.ItemID)
	data.ShareID = types.StringValue(item.ShareID)

	itemType := data.Type.ValueString()
	switch itemType {
	case "login":
		data.Username = setString(item.Username)
		data.Email = setString(item.Email)
		data.Password = setString(item.Password)
		if len(item.URLs) > 0 {
			var urlVals []types.String
			for _, u := range item.URLs {
				urlVals = append(urlVals, types.StringValue(u))
			}
			l, d := types.ListValueFrom(ctx, types.StringType, urlVals)
			diags.Append(d...)
			data.URLs = l
		}
	case "credit-card":
		data.CardholderName = setString(item.CardholderName)
		data.Number = setString(item.Number)
		data.VerificationNumber = setString(item.VerificationNumber)
		data.ExpirationDate = setString(item.ExpirationDate)
		data.PIN = setString(item.PIN)
	case "wifi":
		data.SSID = setString(item.SSID)
		data.Password = setString(item.Password)
		data.Security = setString(item.Security)
	case "ssh-key":
		data.PrivateKey = setString(item.PrivateKey)
		data.PublicKey = setString(item.PublicKey)
	case "identity":
		data.FullName = setString(item.FullName)
		data.Email = setString(item.Email)
		data.PhoneNumber = setString(item.PhoneNumber)
		data.FirstName = setString(item.FirstName)
		data.MiddleName = setString(item.MiddleName)
		data.LastName = setString(item.LastName)
		data.Birthdate = setString(item.Birthdate)
		data.Gender = setString(item.Gender)
		data.Organization = setString(item.Organization)
		data.StreetAddress = setString(item.StreetAddress)
		data.ZipOrPostalCode = setString(item.ZipOrPostalCode)
		data.City = setString(item.City)
		data.StateOrProvince = setString(item.StateOrProvince)
		data.CountryOrRegion = setString(item.CountryOrRegion)
		data.Ssn = setString(item.SocialSecurity)
		data.PassportNumber = setString(item.PassportNumber)
		data.LicenseNumber = setString(item.LicenseNumber)
		data.Website = setString(item.Website)
		data.Company = setString(item.Company)
		data.JobTitle = setString(item.JobTitle)
		data.WorkEmail = setString(item.WorkEmail)
		data.WorkPhoneNumber = setString(item.WorkPhoneNumber)
	}

	data.PasswordWO = types.StringNull()
	data.NoteWO = types.StringNull()
	data.SsnWO = types.StringNull()
	data.PassportNumberWO = types.StringNull()
	data.LicenseNumberWO = types.StringNull()

	return diags
}

func (r *ItemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ItemResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	item, err := r.client.ReadItem(ctx, data.ItemID.ValueString(), data.ShareID.ValueString())
	if err != nil {
		if passcli.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read item", err.Error())
		return
	}

	if item.State == "Trashed" {
		_ = r.client.RestoreItem(ctx, data.ItemID.ValueString(), data.ShareID.ValueString())
		// Fetch again to get updated modify_time and state
		item, _ = r.client.ReadItem(ctx, data.ItemID.ValueString(), data.ShareID.ValueString())
	}

	resp.Diagnostics.Append(r.mapItemToModel(ctx, item, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func getUpdatedSecret(ctx context.Context, req resource.UpdateRequest, planVal types.String, stateVal types.String, planVer types.Int64, stateVer types.Int64, woAttr string) string {
	// If version bumped, read from config WriteOnly attribute
	if !planVer.IsNull() && !planVer.IsUnknown() {
		if stateVer.IsNull() || planVer.ValueInt64() != stateVer.ValueInt64() {
			var v types.String
			req.Config.GetAttribute(ctx, path.Root(woAttr), &v)
			if !v.IsNull() && !v.IsUnknown() && v.ValueString() != "" {
				return v.ValueString()
			}
		}
	}
	// Else if standard attribute changed
	if planVal.ValueString() != stateVal.ValueString() {
		return planVal.ValueString()
	}
	return ""
}

func (r *ItemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ItemResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fields := map[string]string{}

	if plan.Title.ValueString() != state.Title.ValueString() {
		fields["title"] = plan.Title.ValueString()
	}

	itemType := plan.Type.ValueString()

	switch itemType {
	case "login":
		if plan.Username.ValueString() != state.Username.ValueString() {
			fields["username"] = plan.Username.ValueString()
		}
		if plan.Email.ValueString() != state.Email.ValueString() {
			fields["email"] = plan.Email.ValueString()
		}
		if pw := getUpdatedSecret(ctx, req, plan.Password, state.Password, plan.PasswordWOVersion, state.PasswordWOVersion, "password_wo"); pw != "" {
			fields["password"] = pw
		}
	case "note":
		if n := getUpdatedSecret(ctx, req, plan.Note, state.Note, plan.NoteWOVersion, state.NoteWOVersion, "note_wo"); n != "" {
			fields["note"] = n
		}
	case "credit-card":
		if plan.CardholderName.ValueString() != state.CardholderName.ValueString() {
			fields["cardholder_name"] = plan.CardholderName.ValueString()
		}
		if plan.Number.ValueString() != state.Number.ValueString() {
			fields["number"] = plan.Number.ValueString()
		}
		if plan.VerificationNumber.ValueString() != state.VerificationNumber.ValueString() {
			fields["verification_number"] = plan.VerificationNumber.ValueString()
		}
		if plan.ExpirationDate.ValueString() != state.ExpirationDate.ValueString() {
			fields["expiration_date"] = plan.ExpirationDate.ValueString()
		}
		if plan.PIN.ValueString() != state.PIN.ValueString() {
			fields["pin"] = plan.PIN.ValueString()
		}
	case "wifi":
		if plan.SSID.ValueString() != state.SSID.ValueString() {
			fields["ssid"] = plan.SSID.ValueString()
		}
		if pw := getUpdatedSecret(ctx, req, plan.Password, state.Password, plan.PasswordWOVersion, state.PasswordWOVersion, "password_wo"); pw != "" {
			fields["password"] = pw
		}
		if plan.Security.ValueString() != state.Security.ValueString() {
			fields["security"] = plan.Security.ValueString()
		}
	case "identity":
		if plan.FullName.ValueString() != state.FullName.ValueString() {
			fields["full_name"] = plan.FullName.ValueString()
		}
		if plan.Email.ValueString() != state.Email.ValueString() {
			fields["email"] = plan.Email.ValueString()
		}
		if plan.PhoneNumber.ValueString() != state.PhoneNumber.ValueString() {
			fields["phone_number"] = plan.PhoneNumber.ValueString()
		}
		if plan.FirstName.ValueString() != state.FirstName.ValueString() {
			fields["first_name"] = plan.FirstName.ValueString()
		}
		if plan.MiddleName.ValueString() != state.MiddleName.ValueString() {
			fields["middle_name"] = plan.MiddleName.ValueString()
		}
		if plan.LastName.ValueString() != state.LastName.ValueString() {
			fields["last_name"] = plan.LastName.ValueString()
		}
		if plan.Birthdate.ValueString() != state.Birthdate.ValueString() {
			fields["birthdate"] = plan.Birthdate.ValueString()
		}
		if plan.Gender.ValueString() != state.Gender.ValueString() {
			fields["gender"] = plan.Gender.ValueString()
		}
		if plan.Organization.ValueString() != state.Organization.ValueString() {
			fields["organization"] = plan.Organization.ValueString()
		}
		if plan.StreetAddress.ValueString() != state.StreetAddress.ValueString() {
			fields["street_address"] = plan.StreetAddress.ValueString()
		}
		if plan.ZipOrPostalCode.ValueString() != state.ZipOrPostalCode.ValueString() {
			fields["zip_or_postal_code"] = plan.ZipOrPostalCode.ValueString()
		}
		if plan.City.ValueString() != state.City.ValueString() {
			fields["city"] = plan.City.ValueString()
		}
		if plan.StateOrProvince.ValueString() != state.StateOrProvince.ValueString() {
			fields["state_or_province"] = plan.StateOrProvince.ValueString()
		}
		if plan.CountryOrRegion.ValueString() != state.CountryOrRegion.ValueString() {
			fields["country_or_region"] = plan.CountryOrRegion.ValueString()
		}
		if plan.Website.ValueString() != state.Website.ValueString() {
			fields["website"] = plan.Website.ValueString()
		}
		if plan.Company.ValueString() != state.Company.ValueString() {
			fields["company"] = plan.Company.ValueString()
		}
		if plan.JobTitle.ValueString() != state.JobTitle.ValueString() {
			fields["job_title"] = plan.JobTitle.ValueString()
		}
		if plan.WorkEmail.ValueString() != state.WorkEmail.ValueString() {
			fields["work_email"] = plan.WorkEmail.ValueString()
		}
		if plan.WorkPhoneNumber.ValueString() != state.WorkPhoneNumber.ValueString() {
			fields["work_phone_number"] = plan.WorkPhoneNumber.ValueString()
		}

		if ssn := getUpdatedSecret(ctx, req, plan.Ssn, state.Ssn, plan.SsnWOVersion, state.SsnWOVersion, "ssn_wo"); ssn != "" {
			fields["social_security_number"] = ssn
		}
		if pn := getUpdatedSecret(ctx, req, plan.PassportNumber, state.PassportNumber, plan.PassportNumberWOVersion, state.PassportNumberWOVersion, "passport_number_wo"); pn != "" {
			fields["passport_number"] = pn
		}
		if ln := getUpdatedSecret(ctx, req, plan.LicenseNumber, state.LicenseNumber, plan.LicenseNumberWOVersion, state.LicenseNumberWOVersion, "license_number_wo"); ln != "" {
			fields["license_number"] = ln
		}
	}

	if len(fields) > 0 {
		err := r.client.UpdateItem(ctx, state.ItemID.ValueString(), state.ShareID.ValueString(), fields)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update item", err.Error())
			return
		}
	}

	item, err := r.client.ReadItem(ctx, state.ItemID.ValueString(), state.ShareID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read item after update", err.Error())
		return
	}

	plan.ModifyTime = types.StringValue(item.ModifyTime)

	resp.Diagnostics.Append(r.mapItemToModel(ctx, item, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ItemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ItemResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.DeleteItem(ctx, data.ItemID.ValueString(), data.ShareID.ValueString(), data.DestroyPermanently.ValueBool())
	if err != nil {
		// Ignore if already trashed or not found
		msg := err.Error()
		if passcli.IsNotFound(err) ||
			strings.Contains(msg, "Current state: Trashed") ||
			strings.Contains(msg, "Item not found") {
			return
		}
		resp.Diagnostics.AddError("Failed to delete item", err.Error())
		return
	}
}

func (r *ItemResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// ImportState splits formatted ID "share_id:item_id" to populate resource
	idParts := strings.Split(req.ID, ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError("Unexpected Import Identifier", "Expected format: share_id:item_id")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("share_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("item_id"), idParts[1])...)
}
