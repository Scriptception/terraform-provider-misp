resource "misp_sharing_group" "partners" {
  name          = "Partner Network"
  releasability = "Releasable to named partners only"
  active        = true
}

resource "misp_organisation" "acme" {
  name  = "ACME"
  local = true
}

# Add ACME as a member of the Partner Network sharing group.
resource "misp_sharing_group_member" "acme_partner" {
  sharing_group_id = misp_sharing_group.partners.id
  organisation_id  = misp_organisation.acme.id
}
