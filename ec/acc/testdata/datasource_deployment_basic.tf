resource "ec_deployment" "basic_datasource" {
  name    = "%s"
  region  = "%s"
  version = "%s"

  # TODO: Make this template ID dependent on the region.
  # This test should be the only one which uses the 
  # "aws-compute-optimized-v2" template in order to have
  # consistent query results.
  deployment_template_id = "aws-compute-optimized-v2"

  elasticsearch {
    topology {
      instance_configuration_id = "aws.data.highcpu.m5d"
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

data "ec_deployment" "success" {
  id = ec_deployment.basic_datasource.id
}

data "ec_deployments" "query" {
  name_prefix            = substr(ec_deployment.basic_datasource.name, 0, 22)
  deployment_template_id = "aws-compute-optimized-v2"

  elasticsearch {
    version = "%s"
  }

  kibana {
    version = "%s"
  }

  apm {
    version = "%s"
  }

  enterprise_search {
    version = "%s"
  }
}