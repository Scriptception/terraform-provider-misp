resource "misp_role" "analyst" {
  name       = "Analyst"
  permission = "2" # add + modify + modify_org

  perm_tagger        = true
  perm_sharing_group = true
  perm_sighting      = true
  perm_template      = false
  perm_sync          = false
  perm_admin         = false
  perm_site_admin    = false
}
