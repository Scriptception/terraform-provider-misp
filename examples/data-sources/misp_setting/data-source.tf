data "misp_setting" "baseurl" {
  name = "MISP.baseurl"
}

output "misp_baseurl" {
  value = data.misp_setting.baseurl.value
}
