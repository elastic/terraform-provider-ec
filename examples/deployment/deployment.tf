terraform {
  # The Elastic Cloud provider is supported from ">=0.12"
  # Version later than 0.12.29 is required for this terraform block to work.
  required_version = ">= 0.12.29"

  required_providers {
    ec = {
      source  = "elastic/ec"
      version = "0.3.0"
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

  elasticsearch {
    config {
      user_settings_yaml = file("./es_settings.yaml")
    }
  }

  kibana {}

  enterprise_search {
    topology {
      zone_count = 1
    }
  }

  apm {
    topology {
      size = "0.5g"
    }
  }
}