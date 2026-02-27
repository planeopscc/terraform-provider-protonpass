data "protonpass_item_note" "example" {
  item_id  = "item_id_here"
  share_id = "share_id_here"
}

output "note_content" {
  value     = data.protonpass_item_note.example.note
  sensitive = true
}
