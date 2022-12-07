data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "basic" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch = {
    hot = {
      size        = "1g"
      autoscaling = {}
    }
  }

  kibana = {}

  apm = {}

  enterprise_search = {}

  traffic_filter = [
    ec_deployment_traffic_filter.default.id,
  ]
}

resource "ec_deployment_traffic_filter" "default" {
  name   = "%s"
  region = "%s"
  type   = "ip"

  rule {
    source = "0.0.0.0/0"
  }
}

