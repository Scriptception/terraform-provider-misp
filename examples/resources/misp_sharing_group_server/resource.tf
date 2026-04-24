resource "misp_sharing_group" "partners" {
  name          = "Partner Network"
  releasability = "Releasable to named partners only"
  active        = true
}

resource "misp_server" "peer" {
  name          = "Peer MISP Instance"
  url           = "https://misp-peer.example.com"
  authkey       = var.peer_authkey
  remote_org_id = "1"
  self_signed   = true
}

# Supply via TF_VAR_peer_authkey, -var, *.tfvars (kept out of git), or a
# secrets backend (Vault / SOPS / Doppler / etc.). Never commit the value.
variable "peer_authkey" {
  description = "API key issued by the remote MISP for this sync user."
  type        = string
  sensitive   = true
}

# Add the peer MISP server to the Partner Network sharing group.
# Note: server_id="0" (MISP's auto-managed local-instance entry) is reserved and
# cannot be managed via this resource.
resource "misp_sharing_group_server" "peer_in_partners" {
  sharing_group_id = misp_sharing_group.partners.id
  server_id        = misp_server.peer.id
}
