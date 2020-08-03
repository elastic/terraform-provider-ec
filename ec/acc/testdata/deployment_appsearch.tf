resource "ec_deployment" "appsearch" {
  name                   = "%s"
  region                 = "%s"
  version                = "%s"
  
  # TODO: Make this template ID dependent on the region.
  deployment_template_id = "aws-appsearch-dedicated-v2"

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

  appsearch {
    topology {
      instance_configuration_id = "aws.appsearch.m5d"
    }
  }
}