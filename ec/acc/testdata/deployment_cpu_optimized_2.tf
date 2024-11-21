data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "cpu_optimized" {
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

  apm = {
    size = "2g"
  }
}
