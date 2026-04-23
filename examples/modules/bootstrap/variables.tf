variable "local_org" {
  description = "The local organisation that owns events on this MISP instance."
  type = object({
    name        = string
    description = optional(string)
    type        = optional(string, "CSIRT")
    nationality = optional(string)
    sector      = optional(string)
  })
}

variable "admin_email" {
  description = "Email of the initial site-admin user to create."
  type        = string
}

variable "admin_role_id" {
  description = <<-EOT
    MISP role id for the admin user. Defaults to "1" (site-admin) shipped with MISP.
    Override only if you are using a custom role created elsewhere.
  EOT
  type        = string
  default     = "1"
}

variable "enable_taxonomies" {
  description = <<-EOT
    List of bundled taxonomy namespaces to enable (e.g. ["tlp", "admiralty-scale"]).
    See MISP → Event Actions → List Taxonomies for available namespaces.
  EOT
  type        = list(string)
  default     = ["tlp", "admiralty-scale", "workflow"]
}

variable "enable_warninglists" {
  description = "List of bundled warninglist names to enable (exact name match)."
  type        = list(string)
  default     = []
}

variable "enable_noticelists" {
  description = "List of bundled noticelist names to enable (exact name match)."
  type        = list(string)
  default     = []
}

variable "custom_tags" {
  description = "Custom tags to create in the tag catalog. Keys are Terraform-local identifiers."
  type = map(object({
    name       = string
    colour     = optional(string, "#1f77b4")
    exportable = optional(bool, true)
  }))
  default = {}
}

variable "settings" {
  description = <<-EOT
    Global MISP settings to enforce (e.g. { "MISP.baseurl" = "https://misp.example.com" }).
    Leave empty to inherit whatever MISP was configured with out of band.
  EOT
  type        = map(string)
  default     = {}
}
