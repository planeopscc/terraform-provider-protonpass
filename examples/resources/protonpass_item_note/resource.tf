resource "protonpass_item_note" "example" {
  share_id        = protonpass_vault.example.share_id
  title           = "My Note"
  note_wo         = "content"
  note_wo_version = 1
}
