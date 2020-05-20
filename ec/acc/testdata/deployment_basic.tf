resource "ec_deployment" "testacc" {
  name                   = "%s"
  region                 = "%s"
  version                = "%s"
  
  # TODO: Make this template ID dependent on the region.
  deployment_template_id = "aws-io-optimized"

  elasticsearch {
    topology {
      instance_configuration_id = "aws.data.highio.i3"
      memory_per_node = "1g"
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