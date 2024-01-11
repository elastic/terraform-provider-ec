data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "compute_optimized" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch = {
    hot = {
      instance_configuration_id = "aws.es.datahot.m5d"
      autoscaling               = {}
    }
  }

  kibana = {}
}