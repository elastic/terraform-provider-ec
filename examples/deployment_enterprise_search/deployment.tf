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
  name = "my_example_deployment"

  # Mandatory fields
  region                 = "us-east-1"
  version                = "7.8.1"
  deployment_template_id = "aws-enterprise-search-dedicated-v2"

  elasticsearch {
    topology {
      instance_configuration_id = "aws.data.highcpu.m5d"
      memory_per_node = "1g"
    }
  }

  kibana {
    topology {
      instance_configuration_id = "aws.kibana.r5d"
    }
  }

  enterprise_search {
    topology {
      instance_configuration_id = "aws.enterprisesearch.m5d"
    }
  }
}