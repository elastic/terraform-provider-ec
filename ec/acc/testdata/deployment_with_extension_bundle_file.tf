locals {
  region              = "%s"
  deployment_template = "%s"
  name                = "%s"
  description         = "%s"
  file_path           = "%s"
}

data "ec_stack" "latest" {
  version_regex = "latest"
  region        = local.region
}

resource "ec_deployment" "with_extension" {
  name                   = local.name
  region                 = local.region
  version                = data.ec_stack.latest.version
  deployment_template_id = local.deployment_template

  elasticsearch = {
    hot = {
      autoscaling = {}
    }
    extension = [{
      type    = "bundle"
      name    = local.name
      version = data.ec_stack.latest.version
      url     = ec_deployment_extension.my_extension.url
    }]
  }
}

resource "ec_deployment_extension" "my_extension" {
  name           = local.name
  description    = local.description
  version        = "*"
  extension_type = "bundle"

  file_path = local.file_path
  file_hash = filebase64sha256(local.file_path)
}
