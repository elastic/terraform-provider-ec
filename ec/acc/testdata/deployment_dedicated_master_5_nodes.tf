data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "auto_dedicated_master" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch = {
    hot = {
      size        = "2g"
      zone_count  = 2
      autoscaling = {}
    }

    warm = {
      size        = "2g"
      zone_count  = 3
      autoscaling = {}
    }
  }

  kibana = {}
}
