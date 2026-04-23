data "misp_warninglist" "akamai" {
  name = "List of known Akamai IP ranges"
}

output "akamai_enabled" {
  value = data.misp_warninglist.akamai.enabled
}
