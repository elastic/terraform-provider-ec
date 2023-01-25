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
    topology = {
      "hot_content" = {
        size        = "1g"
        autoscaling = {}
      }

      "warm" = {
        autoscaling = {}
      }

      "cold" = {
        autoscaling = {}
      }

      "frozen" = {
        autoscaling = {}
      }

      "ml" = {
        autoscaling = {}
      }

      "master" = {
        autoscaling = {}
      }

      "coordinating" = {
        autoscaling = {}
      }
    }
  }

  kibana = {
    instance_configuration_id = "%s"
  }

  apm = {
    instance_configuration_id = "%s"
  }

  enterprise_search = {
    instance_configuration_id = "%s"
  }
}
