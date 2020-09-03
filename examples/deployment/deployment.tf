terraform {
  required_version = ">= 0.12.29"

  required_providers {
    ec = {
      source = "elastic/ec"
    }
  }
}

provider "ec" {}

# Create an Elastic Cloud deployment
resource "ec_deployment" "example_minimal" {
  # Optional name.
  name = "my_example_deployment"

  # Mandatory fields
  region                 = "us-east-1"
  version                = "7.8.1"
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch {
    topology {
      instance_configuration_id = "aws.data.highio.i3"
    }
  }

  kibana {
    topology {
      instance_configuration_id = "aws.kibana.r5d"
    }
  }

  apm {
    topology {
      instance_configuration_id = "aws.apm.r5d"
    }
  }

  enterprise_search {
    topology {
      instance_configuration_id = "aws.enterprisesearch.m5d"
    }
  }
}