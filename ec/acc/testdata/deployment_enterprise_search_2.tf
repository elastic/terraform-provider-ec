data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "enterprise_search" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch = {
    hot = {
      size        = "2g"
      autoscaling = {}
    }
  }

  kibana = {}
}