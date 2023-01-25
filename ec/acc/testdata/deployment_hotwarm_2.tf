data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "hotwarm" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch = {
    topology = {
      "hot_content" = {
        zone_count  = 1
        size        = "1g"
        autoscaling = {}
      }

      "warm" = {
        zone_count  = 1
        size        = "2g"
        autoscaling = {}
      }
    }
  }
}
