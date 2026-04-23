terraform {
  required_providers {
    misp = {
      source  = "Scriptception/misp"
      version = "~> 0.1"
    }
  }
}

provider "misp" {
  url     = "https://misp.example.com"
  api_key = var.misp_api_key
}

variable "misp_api_key" {
  type      = string
  sensitive = true
}
