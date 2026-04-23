data "misp_organisation" "admin" {
  id = "1"
}

output "admin_name" {
  value = data.misp_organisation.admin.name
}
