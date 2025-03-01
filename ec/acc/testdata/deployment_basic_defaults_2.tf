data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "defaults" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch = {
    hot = {
      autoscaling = {}
    }

    strategy = "rolling_all"
  }

  kibana = {
    size = "2g"
  }

  integrations_server = {
    size = "2g"
  }

  enterprise_search = {
    zone_count = 1
  }
}
