resource "ec_deployment" "testacc" {
  name                   = "%s"
  region                 = "%s"
  version                = "%s"
  
  # TODO: Make this template ID dependent on the region.
  deployment_template_id = "aws-appsearch-dedicated"

  elasticsearch {
    topology {
      instance_configuration_id = "aws.data.highcpu.m5"
      memory_per_node = "1g"
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