resource "misp_organisation" "acme" {
  name        = "ACME"
  description = "ACME Corp threat intel team"
  type        = "CSIRT"
  nationality = "AU"
  sector      = "Technology"
  local       = true
}
