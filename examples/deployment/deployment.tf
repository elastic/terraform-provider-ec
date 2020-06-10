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
  version                = "7.7.0"
  deployment_template_id = "aws-io-optimized"

  elasticsearch {
    topology {
      instance_configuration_id = "aws.data.highio.i3"
    }
  }

  kibana {
    topology {
      instance_configuration_id = "aws.kibana.r4"
    }
  }

  apm {
    topology {
      instance_configuration_id = "aws.apm.r4"
    }
  }
}