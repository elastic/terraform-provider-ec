data "ec_stack" "autoscaling" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "autoscaling" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.autoscaling.version
  deployment_template_id = "%s"

  elasticsearch = {
    autoscale = "false"

    cold = {
      size        = "0g"
      zone_count  = 1
      autoscaling = {}
    }

    frozen = {
      size        = "0g"
      zone_count  = 1
      autoscaling = {}
    }

    hot = {
      size       = "1g"
      zone_count = 1
      autoscaling = {
        max_size = "8g"
      }
    }

    ml = {
      size       = "0g"
      zone_count = 1
      autoscaling = {
        min_size = "0g"
        max_size = "4g"
      }
    }

    warm = {
      size       = "2g"
      zone_count = 1
      autoscaling = {
        max_size = "15g"
      }
    }
  }
}
