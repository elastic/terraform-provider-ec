terraform {
  required_version = ">= 1.0"

  required_providers {
    ec = {
      source  = "elastic/ec"
      version = "0.6.0"
    }
  }
}

provider "ec" {}

locals {
  file_path = "./files/content.json.zip"
}

# Create an Elastic Cloud Extension
resource "ec_deployment_extension" "example_extension" {
  name           = "my_extension"
  description    = "my extension"
  version        = "*"
  extension_type = "bundle"

  file_path = local.file_path
  file_hash = filebase64sha256(local.file_path)
}
