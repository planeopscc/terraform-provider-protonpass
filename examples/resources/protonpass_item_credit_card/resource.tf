resource "protonpass_item_credit_card" "example" {
  share_id               = protonpass_vault.example.share_id
  title                  = "My Card"
  cardholder_name        = "John"
  card_number_wo         = "4111111111111111"
  card_number_wo_version = 1
  cvv_wo                 = "123"
  cvv_wo_version         = 1
  expiration_date        = "2027-12"
}
