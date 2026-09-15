data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "cpu_optimized" {
  name                       = "%s"
  region                     = "%s"
  version                    = data.ec_stack.latest.version
  deployment_template_id     = "%s"
  migrate_to_latest_hardware = true

  elasticsearch = {
    hot = {
      autoscaling = {}
    }
  }

  kibana = {}
}
