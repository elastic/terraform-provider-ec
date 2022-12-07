data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "basic_observability" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch = {
    hot = {
      size        = "1g"
      zone_count  = 1
      autoscaling = {}
    }
  }
}

resource "ec_deployment" "basic_datasource" {
  alias                  = "%s"
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch = {
    hot = {
      size        = "1g"
      zone_count  = 1
      autoscaling = {}
    }
  }

  kibana = {}

  apm = {}

  enterprise_search = {}

  observability = {
    deployment_id = ec_deployment.basic_observability.id
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
  deployment_template_id = "%s"

  elasticsearch {
    version = data.ec_stack.latest.version
  }

  kibana {
    version = data.ec_stack.latest.version
  }

  apm {
    version = data.ec_stack.latest.version
  }

  enterprise_search {
    version = data.ec_stack.latest.version
  }

  depends_on = [
    ec_deployment.basic_datasource,
  ]
}