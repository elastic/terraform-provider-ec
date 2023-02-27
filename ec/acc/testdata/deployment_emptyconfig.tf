data "ec_stack" "emptyconfig" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "emptyconfig" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.emptyconfig.version
  deployment_template_id = "%s"

  elasticsearch = {
    config = {
      user_settings_yaml = null
    }
    hot = {
      size        = "1g"
      zone_count  = 1
      autoscaling = {}
    }
  }
}