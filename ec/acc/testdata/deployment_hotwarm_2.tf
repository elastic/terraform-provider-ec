resource "ec_deployment" "hotwarm" {
  name    = "%s"
  region  = "%s"
  version = "%s"

  # TODO: Make this template ID dependent on the region.
  deployment_template_id = "aws-hot-warm-v2"

  elasticsearch {
    topology {
      instance_configuration_id = "aws.data.highio.i3"
      zone_count                = 1
      size                      = "1g"
    }
    topology {
      instance_configuration_id = "aws.data.highstorage.d2"
      zone_count                = 1
      size                      = "2g"
    }
  }
}