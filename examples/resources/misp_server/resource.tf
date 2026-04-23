resource "misp_server" "peer" {
  name          = "MISP Peer"
  url           = "https://misp-peer.example.com"
  authkey       = "0123456789abcdef0123456789abcdef01234567"
  remote_org_id = "2"

  push          = true
  pull          = true
  self_signed   = false
}
