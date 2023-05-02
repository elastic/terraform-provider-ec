locals {
  file_path = "/path/to/plugin.zip"
}

resource "ec_deployment_extension" "example_extension" {
  name           = "my_extension"
  description    = "my extension"
  version        = "*"
  extension_type = "bundle"

  file_path = local.file_path
  file_hash = filebase64sha256(local.file_path)
}

data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
}

resource "ec_deployment" "with_extension" {
  # Optional name.
  name = "my_example_deployment"

  # Mandatory fields
  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch = {
    hot = {
      autoscaling = {}
    }
    extension = [{
      name    = ec_deployment_extension.example_extension.name
      type    = "bundle"
      version = data.ec_stack.latest.version
      url     = ec_deployment_extension.example_extension.url
    }]
  }
}
