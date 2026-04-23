data "misp_taxonomy" "tlp" {
  namespace = "tlp"
}

output "tlp_enabled" {
  value = data.misp_taxonomy.tlp.enabled
}
