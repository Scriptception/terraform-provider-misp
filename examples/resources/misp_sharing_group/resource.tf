resource "misp_sharing_group" "partners" {
  name          = "Partner Network"
  description   = "Shared with trusted partners"
  releasability = "Releasable to named partners only"
  active        = true
  local         = false
  roaming       = false
}
