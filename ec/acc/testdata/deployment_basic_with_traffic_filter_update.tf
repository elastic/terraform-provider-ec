data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "basic" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch {
    topology {
      size = "1g"
    }
  }

  kibana {}

  apm {}

  enterprise_search {}

  traffic_filter = [
    ec_deployment_traffic_filter.second.id,
  ]
}

resource "ec_deployment_traffic_filter" "second" {
  name   = "%s"
  region = "%s"
  type   = "ip"

  rule {
    source = "0.0.0.0/0"
  }
}

