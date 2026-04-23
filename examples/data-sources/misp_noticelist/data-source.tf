data "misp_noticelist" "gdpr" {
  name = "gdpr"
}

output "gdpr_enabled" {
  value = data.misp_noticelist.gdpr.enabled
}
