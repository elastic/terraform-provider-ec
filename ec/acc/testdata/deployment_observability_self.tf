data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "observability" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  observability = {
    deployment_id = "self"
  }

  elasticsearch = {
    autoscale = "false"

    hot = {
      size        = "1g"
      zone_count  = 1
      autoscaling = {}
    }
  }

  kibana = {
    size       = "1g"
    zone_count = 1
  }
}