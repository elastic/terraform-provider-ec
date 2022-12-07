terraform {
  # The Elastic Cloud provider is supported from ">=0.12"
  # Version later than 0.12.29 is required for this terraform block to work.
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

# Create an Elastic Cloud deployment
resource "ec_deployment" "example_minimal" {
  # Optional name.
  name = "my_example_deployment"

  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch = {

    hot = {
      autoscaling = {}
    }
    config = {
      user_settings_yaml = file("./es_settings.yaml")
    }
  }

  kibana = {}

  enterprise_search = {
    zone_count = 1
  }

  apm = {
    size = "0.5g"
  }
}