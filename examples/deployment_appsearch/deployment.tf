terraform {
  required_version = ">= 0.12"
}

provider "ec" {
}

# Create an Elastic Cloud deployment
resource "ec_deployment" "example_minimal" {
  # Optional name.
  name = "my_example_deployment"

  # Mandatory fields
  region                 = "us-east-1"
  version                = "7.6.2"
  deployment_template_id = "aws-appsearch-dedicated"

  elasticsearch {
    topology {
      instance_configuration_id = "aws.data.highcpu.m5"
    }
  }

  kibana {
    topology {
      instance_configuration_id = "aws.kibana.r4"
    }
  }

  appsearch {
    topology {
      instance_configuration_id = "aws.appsearch.m5"
    }
  }
}