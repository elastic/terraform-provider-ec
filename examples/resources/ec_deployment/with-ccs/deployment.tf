data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
}

resource "ec_deployment" "source_deployment" {
  name = "my_ccs_source"

  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch = {
    hot = {
      size        = "1g"
      autoscaling = {}
    }
  }
}

resource "ec_deployment" "ccs" {
  name = "ccs deployment"

  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-cross-cluster-search-v2"

  elasticsearch = {
    hot = {
      autoscalign = {}
    }
    remote_cluster = [{
      deployment_id = ec_deployment.source_deployment.id
      alias         = ec_deployment.source_deployment.name
      ref_id        = ec_deployment.source_deployment.elasticsearch.0.ref_id
    }]
  }

  kibana = {}
}
