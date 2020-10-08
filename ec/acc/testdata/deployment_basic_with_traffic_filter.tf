resource "ec_deployment" "basic" {
  name    = "%s"
  region  = "%s"
  version = "%s"

  # TODO: Make this template ID dependent on the region.
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch {
    topology {
      instance_configuration_id = "aws.data.highio.i3"
      size                      = "1g"
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

  traffic_filter = [
    ec_deployment_traffic_filter.default.id,
  ]
}

resource "ec_deployment_traffic_filter" "default" {
  name   = "%s"
  region = "%s"
  type   = "ip"

  rule {
    source = "0.0.0.0/0"
  }
}

