resource "ec_deployment" "defaults" {
  name    = "%s"
  region  = "%s"
  version = "%s"

  # TODO: Make this template ID dependent on the region.
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch {}

  kibana {}

  enterprise_search {
    topology {
      zone_count = 1
    }
  }
}