data "ec_stack" "keystore" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "keystore" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.keystore.version
  deployment_template_id = "%s"

  elasticsearch = {
    hot = {
      size        = "1g"
      zone_count  = 1
      autoscaling = {}
    }
  }
}
