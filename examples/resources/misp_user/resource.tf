resource "misp_user" "analyst" {
  email    = "alice@example.com"
  org_id   = "1"
  role_id  = "3" # user
  disabled = false
}
