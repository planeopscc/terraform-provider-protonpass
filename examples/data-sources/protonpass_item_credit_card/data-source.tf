data "protonpass_item_credit_card" "example" {
  item_id  = "item_id_here"
  share_id = "share_id_here"
}

output "cardholder" {
  value = data.protonpass_item_credit_card.example.cardholder_name
}
