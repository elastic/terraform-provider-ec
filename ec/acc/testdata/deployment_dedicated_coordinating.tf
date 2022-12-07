data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "dedicated_coordinating" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch = {
    coordinating = {
      zone_count  = 2
      size        = "1g"
      autoscaling = {}
    }

    hot = {
      zone_count  = 1
      size        = "1g"
      autoscaling = {}
    }

    warm = {
      zone_count  = 1
      size        = "2g"
      autoscaling = {}
    }
  }
}