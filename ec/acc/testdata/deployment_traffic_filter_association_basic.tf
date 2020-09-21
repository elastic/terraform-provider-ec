resource "ec_deployment" "tf_assoc" {
  name    = "%s"
  region  = "%s"
  version = "%s"

  # TODO: Make this template ID dependent on the region.
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch {
    topology {
      instance_configuration_id = "aws.data.highio.i3"
      memory_per_node           = "1g"
    }
  }

  kibana {
    topology {
      instance_configuration_id = "aws.kibana.r5d"
    }
  }
}

resource "ec_deployment_traffic_filter" "tf_assoc" {
  name   = "%s"
  region = "%s"
  type   = "ip"

  rule {
    source = "0.0.0.0/0"
  }
}

resource "ec_deployment_traffic_filter_association" "tf_assoc" {
  traffic_filter_id = ec_deployment_traffic_filter.tf_assoc.id
  deployment_id     = ec_deployment.tf_assoc.id
}
