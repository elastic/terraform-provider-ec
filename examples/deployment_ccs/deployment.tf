terraform {
  required_version = ">= 1.0"

  required_providers {
    ec = {
      source  = "elastic/ec"
      version = "0.6.0"
    }
  }
}

provider "ec" {}

# Retrieve the latest stack pack version
data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
}

resource "ec_deployment" "source_deployment" {
  name = "my_source_ccs"

  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch = {
    config = {}
    hot = {
      zone_count  = 1
      size        = "2g"
      autoscaling = {}
    }
  }
}

resource "ec_deployment" "second_source" {
  name = "my_second_source_source_ccs"

  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch = {
    config = {}
    hot = {
      zone_count  = 1
      size        = "2g"
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
    config = {}
    hot = {
      autoscaling = {}
    }

    remote_cluster = [
      {
        deployment_id = ec_deployment.source_deployment.id
        alias         = ec_deployment.source_deployment.name
        ref_id        = ec_deployment.source_deployment.elasticsearch.0.ref_id
      },
      {
        deployment_id = ec_deployment.second_source.id
        alias         = ec_deployment.second_source.name
        ref_id        = ec_deployment.second_source.elasticsearch.0.ref_id
      }
    ]
  }

  kibana = {}
}
