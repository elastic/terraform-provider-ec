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
    topology = {
      "hot_content" = {
        size        = "2g"
        autoscaling = {}
      }
    }
  }

  kibana = {}

  enterprise_search = {}
}
