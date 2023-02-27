locals {
  region              = "%s"
  deployment_template = "%s"
}

data "ec_stack" "latest" {
  version_regex = "latest"
  region        = local.region
}

resource "ec_deployment" "snapshot_source" {
  name                   = "terraform_acc_snapshot_source"
  region                 = local.region
  version                = data.ec_stack.latest.version
  deployment_template_id = local.deployment_template

  elasticsearch = {
    hot = {
      size        = "1g"
      autoscaling = {}
    }
  }
}

resource "ec_deployment" "snapshot_target" {
  name                   = "terraform_acc_snapshot_target"
  region                 = local.region
  version                = data.ec_stack.latest.version
  deployment_template_id = local.deployment_template

  elasticsearch = {

    snapshot_source = [{
      source_elasticsearch_cluster_id = ec_deployment.snapshot_source.elasticsearch.0.resource_id
    }]

    hot = {
      size        = "1g"
      autoscaling = {}
    }
  }
}
