// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package passcli

// VaultJSON represents a vault as returned by pass-cli.
type VaultJSON struct {
	VaultID string `json:"vault_id"`
	ShareID string `json:"share_id"`
	Name    string `json:"name"`
}

// VaultMemberJSON represents a vault member as returned by pass-cli vault member list --output json.
type VaultMemberJSON struct {
	MemberShareID string `json:"member_share_id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	IsGroupShare  bool   `json:"is_group_share"`
	Role          string `json:"role"`
	TargetType    string `json:"target_type"`
}

// AliasCreateJSON represents the response from pass-cli item alias create --output json.
type AliasCreateJSON struct {
	ID    string `json:"id"`
	Alias string `json:"alias"`
}

// ItemTOTPJSON represents a TOTP code returned by pass-cli item totp --output json.
type ItemTOTPJSON struct {
	Code string `json:"code"`
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

// NoteContentJSON holds note-specific content.
type NoteContentJSON struct {
	Note string `json:"note"` // sometimes empty string or null, main note is in ItemContentJSON
}

// CreditCardContentJSON holds credit card fields.
type CreditCardContentJSON struct {
	CardholderName     string `json:"cardholder_name"`
	Number             string `json:"number"`
	VerificationNumber string `json:"verification_number"` // CVV
	ExpirationDate     string `json:"expiration_date"`
	PIN                string `json:"pin"`
}

// WifiContentJSON holds WiFi network fields.
type WifiContentJSON struct {
	SSID     string `json:"ssid"`
	Password string `json:"password"`
	Security string `json:"security"`
}

// SshKeyContentJSON holds SSH key fields.
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

// ItemJSON is a flattened representation used internally by the provider to represent any item type.
type ItemJSON struct {
	ItemID     string
	ShareID    string
	Title      string
	Type       string // login, note, credit-card, wifi, ssh-key, identity, etc.
	Note       string
	CreateTime string
	ModifyTime string
	State      string

	// Login
	Username string
	Email    string
	Password string
	URLs     []string
	TOTPUri  string

	// Credit Card
	CardholderName     string
	Number             string
	VerificationNumber string
	ExpirationDate     string
	PIN                string

	// WiFi
	SSID     string
	Security string

	// SSH Key
	PrivateKey string
	PublicKey  string

	// Identity
	FullName        string
	PhoneNumber     string
	FirstName       string
	MiddleName      string
	LastName        string
	Birthdate       string
	Gender          string
	Organization    string
	StreetAddress   string
	ZipOrPostalCode string
	City            string
	StateOrProvince string
	CountryOrRegion string
	Floor           string
	County          string
	SocialSecurity  string
	PassportNumber  string
	LicenseNumber   string
	Website         string
	XHandle         string
	SecondPhone     string
	LinkedIn        string
	Reddit          string
	Facebook        string
	Yahoo           string
	Instagram       string
	Company         string
	JobTitle        string
	PersonalWebsite string
	WorkPhoneNumber string
	WorkEmail       string
}

// FlattenItem converts a raw CLI item into our flattened internal representation.
func FlattenItem(raw ItemRawJSON) ItemJSON {
	item := ItemJSON{
		ItemID:     raw.ID,
		ShareID:    raw.ShareID,
		Title:      raw.Content.Title,
		Note:       raw.Content.Note,
		CreateTime: raw.CreateTime,
		ModifyTime: raw.ModifyTime,
		State:      raw.State,
	}

	if raw.Content.Content.Login != nil {
		item.Type = "login"
		item.Username = raw.Content.Content.Login.Username
		item.Email = raw.Content.Content.Login.Email
		item.Password = raw.Content.Content.Login.Password
		item.URLs = raw.Content.Content.Login.URLs
		item.TOTPUri = raw.Content.Content.Login.TOTPUri
	} else if raw.Content.Content.Login == nil && raw.Content.Content.CreditCard == nil && raw.Content.Content.Wifi == nil && raw.Content.Content.SshKey == nil && raw.Content.Content.Identity == nil {
		item.Type = "note"
		if item.Note == "" && raw.Content.Content.Note != nil && raw.Content.Content.Note.Note != "" {
			item.Note = raw.Content.Content.Note.Note
		}
	} else if raw.Content.Content.CreditCard != nil {
		item.Type = "credit-card"
		item.CardholderName = raw.Content.Content.CreditCard.CardholderName
		item.Number = raw.Content.Content.CreditCard.Number
		item.VerificationNumber = raw.Content.Content.CreditCard.VerificationNumber
		item.ExpirationDate = raw.Content.Content.CreditCard.ExpirationDate
		item.PIN = raw.Content.Content.CreditCard.PIN
	} else if raw.Content.Content.Wifi != nil {
		item.Type = "wifi"
		item.SSID = raw.Content.Content.Wifi.SSID
		item.Password = raw.Content.Content.Wifi.Password
		item.Security = raw.Content.Content.Wifi.Security
	} else if raw.Content.Content.SshKey != nil {
		item.Type = "ssh-key"
		item.PrivateKey = raw.Content.Content.SshKey.PrivateKey
		item.PublicKey = raw.Content.Content.SshKey.PublicKey
	} else if raw.Content.Content.Identity != nil {
		item.Type = "identity"
		item.FullName = raw.Content.Content.Identity.FullName
		item.Email = raw.Content.Content.Identity.Email
		item.PhoneNumber = raw.Content.Content.Identity.PhoneNumber
		item.FirstName = raw.Content.Content.Identity.FirstName
		item.MiddleName = raw.Content.Content.Identity.MiddleName
		item.LastName = raw.Content.Content.Identity.LastName
		item.Birthdate = raw.Content.Content.Identity.Birthdate
		item.Gender = raw.Content.Content.Identity.Gender
		item.Organization = raw.Content.Content.Identity.Organization
		item.StreetAddress = raw.Content.Content.Identity.StreetAddress
		item.ZipOrPostalCode = raw.Content.Content.Identity.ZipOrPostalCode
		item.City = raw.Content.Content.Identity.City
		item.StateOrProvince = raw.Content.Content.Identity.StateOrProvince
		item.CountryOrRegion = raw.Content.Content.Identity.CountryOrRegion
		item.Floor = raw.Content.Content.Identity.Floor
		item.County = raw.Content.Content.Identity.County
		item.SocialSecurity = raw.Content.Content.Identity.SocialSecurity
		item.PassportNumber = raw.Content.Content.Identity.PassportNumber
		item.LicenseNumber = raw.Content.Content.Identity.LicenseNumber
		item.Website = raw.Content.Content.Identity.Website
		item.XHandle = raw.Content.Content.Identity.XHandle
		item.SecondPhone = raw.Content.Content.Identity.SecondPhone
		item.LinkedIn = raw.Content.Content.Identity.LinkedIn
		item.Reddit = raw.Content.Content.Identity.Reddit
		item.Facebook = raw.Content.Content.Identity.Facebook
		item.Yahoo = raw.Content.Content.Identity.Yahoo
		item.Instagram = raw.Content.Content.Identity.Instagram
		item.Company = raw.Content.Content.Identity.Company
		item.JobTitle = raw.Content.Content.Identity.JobTitle
		item.PersonalWebsite = raw.Content.Content.Identity.PersonalWebsite
		item.WorkPhoneNumber = raw.Content.Content.Identity.WorkPhoneNumber
		item.WorkEmail = raw.Content.Content.Identity.WorkEmail
	}

	return item
}
