# Retrieve all items of a specific type from a vault.

data "protonpass_items" "all_logins" {
  share_id = "share_id_here"
  type     = "login"
}

output "login_titles" {
  value = [for i in data.protonpass_items.all_logins.items : i.title]
}
