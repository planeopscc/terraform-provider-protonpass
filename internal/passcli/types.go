// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package passcli

// VaultJSON represents a vault as returned by pass-cli --output json.
type VaultJSON struct {
	VaultID string `json:"vault_id"`
	ShareID string `json:"share_id"`
	Name    string `json:"name"`
}

// VaultListResponse represents the response from vault list.
type VaultListResponse struct {
	Vaults []VaultJSON `json:"vaults"`
}

// ItemRawJSON represents a raw item as returned by pass-cli item list/view --output json.
// The structure is deeply nested: content.content.Login holds the login-specific fields.
type ItemRawJSON struct {
	ID         string          `json:"id"`
	ShareID    string          `json:"share_id"`
	VaultID    string          `json:"vault_id"`
	Content    ItemContentJSON `json:"content"`
	State      string          `json:"state"`
	CreateTime string          `json:"create_time"`
	ModifyTime string          `json:"modify_time"`
}

// ItemContentJSON wraps the top-level content (title, note) and inner typed content.
type ItemContentJSON struct {
	Title    string           `json:"title"`
	Note     string           `json:"note"`
	ItemUUID string           `json:"item_uuid"`
	Content  ItemTypedContent `json:"content"`
}

// ItemTypedContent holds the type-specific content (Login, Note, etc.).
type ItemTypedContent struct {
	Login      *LoginContentJSON      `json:"Login"`
	Note       *NoteContentJSON       `json:"Note"`
	CreditCard *CreditCardContentJSON `json:"CreditCard"`
	Wifi       *WifiContentJSON       `json:"Wifi"`
	SshKey     *SshKeyContentJSON     `json:"SshKey"`
	Identity   *IdentityContentJSON   `json:"Identity"`
}

// LoginContentJSON holds login-specific fields.
type LoginContentJSON struct {
	Email    string   `json:"email"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	URLs     []string `json:"urls"`
	TOTPUri  string   `json:"totp_uri"`
}

type NoteContentJSON struct {
	Note string `json:"note"` // sometimes empty string or null, main note is in ItemContentJSON
}

type CreditCardContentJSON struct {
	CardholderName     string `json:"cardholder_name"`
	Number             string `json:"number"`
	VerificationNumber string `json:"verification_number"` // CVV
	ExpirationDate     string `json:"expiration_date"`
	PIN                string `json:"pin"`
}

type WifiContentJSON struct {
	SSID     string `json:"ssid"`
	Password string `json:"password"`
	Security string `json:"security"`
}

type SshKeyContentJSON struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

type IdentityContentJSON struct {
	FullName        string `json:"full_name"`
	Email           string `json:"email"`
	PhoneNumber     string `json:"phone_number"`
	FirstName       string `json:"first_name"`
	MiddleName      string `json:"middle_name"`
	LastName        string `json:"last_name"`
	Birthdate       string `json:"birthdate"`
	Gender          string `json:"gender"`
	Organization    string `json:"organization"`
	StreetAddress   string `json:"street_address"`
	ZipOrPostalCode string `json:"zip_or_postal_code"`
	City            string `json:"city"`
	StateOrProvince string `json:"state_or_province"`
	CountryOrRegion string `json:"country_or_region"`
	Floor           string `json:"floor"`
	County          string `json:"county"`
	SocialSecurity  string `json:"social_security_number"`
	PassportNumber  string `json:"passport_number"`
	LicenseNumber   string `json:"license_number"`
	Website         string `json:"website"`
	XHandle         string `json:"x_handle"`
	SecondPhone     string `json:"second_phone_number"`
	LinkedIn        string `json:"linkedin"`
	Reddit          string `json:"reddit"`
	Facebook        string `json:"facebook"`
	Yahoo           string `json:"yahoo"`
	Instagram       string `json:"instagram"`
	Company         string `json:"company"`
	JobTitle        string `json:"job_title"`
	PersonalWebsite string `json:"personal_website"`
	WorkPhoneNumber string `json:"work_phone_number"`
	WorkEmail       string `json:"work_email"`
}

// ItemViewResponse represents the response from item view.
type ItemViewResponse struct {
	Item ItemRawJSON `json:"item"`
}

// ItemListResponse represents the response from item list.
type ItemListResponse struct {
	Items []ItemRawJSON `json:"items"`
}

// ItemLoginJSON is a flattened representation used internally by the provider.
type ItemLoginJSON struct {
	ItemID     string
	ShareID    string
	Title      string
	Username   string
	Email      string
	Note       string
	URLs       []string
	CreateTime string
	ModifyTime string
}

// FlattenItem converts a raw CLI item into our flattened internal representation.
func FlattenItem(raw ItemRawJSON) ItemLoginJSON {
	item := ItemLoginJSON{
		ItemID:     raw.ID,
		ShareID:    raw.ShareID,
		Title:      raw.Content.Title,
		Note:       raw.Content.Note,
		CreateTime: raw.CreateTime,
		ModifyTime: raw.ModifyTime,
	}
	if raw.Content.Content.Login != nil {
		item.Username = raw.Content.Content.Login.Username
		item.Email = raw.Content.Content.Login.Email
		item.URLs = raw.Content.Content.Login.URLs
	}
	return item
}
