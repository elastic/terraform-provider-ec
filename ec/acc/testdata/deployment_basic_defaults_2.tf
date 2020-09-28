resource "ec_deployment" "defaults" {
  name    = "%s"
  region  = "%s"
  version = "%s"

  # TODO: Make this template ID dependent on the region.
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch {
    topology {
      memory_per_node = "1g"
    }
  }

  kibana {
    topology {
      memory_per_node = "2g"
    }
  }

  apm {}

  enterprise_search {}
}