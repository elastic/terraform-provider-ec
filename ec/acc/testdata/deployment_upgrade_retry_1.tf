locals {
  region              = "%s"
  deployment_template = "%s"
}

data "ec_stack" "latest" {
  version_regex = "7.10.?"
  region        = local.region
}

resource "ec_deployment" "upgrade_retry" {
  name                   = "terraform_acc_upgrade_retry"
  region                 = local.region
  version                = data.ec_stack.latest.version
  deployment_template_id = local.deployment_template

  elasticsearch = {
    hot = {
      size        = "1g"
      zone_count  = 1
      autoscaling = {}
    }
  }

  kibana = {}
}
