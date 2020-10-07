resource "ec_deployment" "defaults" {
  name    = "%s"
  region  = "%s"
  version = "%s"

  # TODO: Make this template ID dependent on the region.
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch {}

  kibana {
    topology {
      size = "2g"
    }
  }

  apm {
    topology {
      size = "1g"
    }
  }

  enterprise_search {
    topology {
      zone_count = 1
    }
  }
}