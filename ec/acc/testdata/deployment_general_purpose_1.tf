data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "general_purpose" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch = {
    hot = {
      autoscaling = {}
    }

    warm = {
      autoscaling = {}
    }
  }
}
