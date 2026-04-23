resource "misp_feed" "example" {
  name          = "Abuse.ch URLhaus"
  provider_name = "abuse.ch"
  url           = "https://urlhaus-api.abuse.ch/v1/misp/"
  source_format = "misp"
  enabled       = true
  distribution  = "3"
  input_source  = "network"
}
