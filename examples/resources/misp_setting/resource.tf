# Set the MISP base URL.
# Boolean settings: pass "true" or "false" as a string.
# Numeric settings: pass the number as a string, e.g. "5000".
resource "misp_setting" "baseurl" {
  name  = "MISP.baseurl"
  value = "https://misp.example.com"
}

# Disable outgoing e-mail (boolean setting — value is a string).
resource "misp_setting" "disable_emailing" {
  name  = "MISP.disable_emailing"
  value = "false"
}
