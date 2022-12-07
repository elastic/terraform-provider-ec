data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "dedicated_master" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch = {
    cold = {
      zone_count  = 1
      size        = "2g"
      autoscaling = {}
    }

    hot = {
      zone_count  = 3
      size        = "1g"
      autoscaling = {}
    }

    master = {
      zone_count  = 3
      size        = "1g"
      autoscaling = {}
    }

    warm = {
      zone_count  = 2
      size        = "2g"
      autoscaling = {}
    }
  }
}