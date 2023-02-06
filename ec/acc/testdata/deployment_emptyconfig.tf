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
    topology = {
      "hot_content" = {
        size        = "1g"
        zone_count  = 1
        autoscaling = {}
      }
    }
  }
}
