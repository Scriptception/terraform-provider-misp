# Custom galaxy cluster in the "360net-threat-actor" galaxy (galaxy_id = 1).
# Only custom clusters (default = false) can be managed by Terraform.
# Bundled MISP clusters (MITRE ATT&CK, threat-actor catalogs, etc.) are read-only.

resource "misp_galaxy_cluster" "example" {
  galaxy_id    = "1"
  value        = "My Threat Actor"
  description  = "Internal tracking cluster for a specific threat actor."
  source       = "https://internal.example.com/ta/my-threat-actor"
  authors      = ["security-team"]
  distribution = "1" # This community only

  # Multiple elements may share the same key (e.g. refs, synonyms).
  elements = [
    { key = "refs", value = "https://internal.example.com/ta/my-threat-actor" },
    { key = "refs", value = "https://attack.mitre.org/groups/G0001/" },
    { key = "synonyms", value = "OldActorName" },
  ]
}
