output "local_organisation_id" {
  description = "Numeric MISP id of the local organisation. Reference this when creating sharing groups, users, or feeds outside the module."
  value       = misp_organisation.local.id
}

output "local_organisation_uuid" {
  description = "UUID of the local organisation."
  value       = misp_organisation.local.uuid
}

output "admin_user_id" {
  description = "Numeric MISP id of the admin user created by the module."
  value       = misp_user.admin.id
}

output "custom_tag_ids" {
  description = "Map from tag key (in var.custom_tags) to the tag's numeric MISP id."
  value       = { for k, t in misp_tag.custom : k => t.id }
}
