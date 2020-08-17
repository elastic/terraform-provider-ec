terraform {
  required_version = ">= 0.12.29"
  required_providers {
    ec = {
      source = "elastic/ec"
    }
  }
}

provider "ec" {
}

# Create an Elastic Cloud deployment
resource "ec_deployment" "example_minimal" {
  # Optional name.
  name = "my_appsearch_example_deployment"

  # Mandatory fields
  region                 = "us-east-1"
  version                = "7.6.2"
  deployment_template_id = "aws-appsearch-dedicated-v2"

  elasticsearch {
    topology {
      instance_configuration_id = "aws.data.highcpu.m5d"
    }
  }

  kibana {
    topology {
      instance_configuration_id = "aws.kibana.r5d"
    }
  }

  appsearch {
    topology {
      instance_configuration_id = "aws.appsearch.m5d"
    }
  }
}