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
      size        = "1g"
      autoscaling = {}
    }
  }

  kibana = {}

  integrations_server = {}

  enterprise_search = {}
}
