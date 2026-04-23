terraform {
  required_version = ">= 1.6"
  required_providers {
    misp = {
      source  = "Scriptception/misp"
      version = "~> 0.1"
    }
  }
}

# Local organisation — owns events created on this instance.
resource "misp_organisation" "local" {
  name        = var.local_org.name
  description = var.local_org.description
  type        = var.local_org.type
  nationality = var.local_org.nationality
  sector      = var.local_org.sector
  local       = true
}

# Initial admin user.
resource "misp_user" "admin" {
  email   = var.admin_email
  org_id  = misp_organisation.local.id
  role_id = var.admin_role_id
}

# Enable bundled taxonomies. Uses the adopt-existing pattern — MISP ships these
# disabled by default; this resource toggles `enabled=true`.
resource "misp_taxonomy" "enabled" {
  for_each  = toset(var.enable_taxonomies)
  namespace = each.key
  enabled   = true
}

# Enable bundled warninglists.
resource "misp_warninglist" "enabled" {
  for_each = toset(var.enable_warninglists)
  name     = each.key
  enabled  = true
}

# Enable bundled noticelists.
resource "misp_noticelist" "enabled" {
  for_each = toset(var.enable_noticelists)
  name     = each.key
  enabled  = true
}

# Custom tags.
resource "misp_tag" "custom" {
  for_each   = var.custom_tags
  name       = each.value.name
  colour     = each.value.colour
  exportable = each.value.exportable
}

# Global settings.
resource "misp_setting" "enforced" {
  for_each = var.settings
  name     = each.key
  value    = each.value
}
