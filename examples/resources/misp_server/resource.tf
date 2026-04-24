resource "misp_server" "peer" {
  name          = "MISP Peer"
  url           = "https://misp-peer.example.com"
  authkey       = var.peer_authkey
  remote_org_id = "2"

  push        = true
  pull        = true
  self_signed = false
}

# The remote MISP's API key. Never hardcode this — supply via:
#   TF_VAR_peer_authkey=...   (environment)
#   terraform apply -var "peer_authkey=..."
#   a *.tfvars file kept out of git
#   or a secrets backend (Vault / SOPS / Doppler / etc.)
variable "peer_authkey" {
  description = "API key issued by the remote MISP for this sync user."
  type        = string
  sensitive   = true
}
