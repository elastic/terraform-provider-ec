resource "ec_deployment" "defaults" {
  name    = "%s"
  region  = "%s"
  version = "%s"

  # TODO: Make this template ID dependent on the region.
  deployment_template_id = "aws-hot-warm-v2"

  elasticsearch {}
}