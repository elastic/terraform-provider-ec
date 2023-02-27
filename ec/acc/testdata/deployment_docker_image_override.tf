locals {
  name                = "%s"
  region              = "%s"
  deployment_template = "%s"
}

data "ec_stack" "latest" {
  version_regex = "7.15.?"
  region        = local.region
}

resource "ec_deployment" "docker_image" {
  name                   = local.name
  region                 = local.region
  version                = data.ec_stack.latest.version
  deployment_template_id = local.deployment_template

  elasticsearch = {
    config = {
      docker_image = "docker.elastic.co/cloud-ci/elasticsearch:7.15.0-SNAPSHOT"
    }

    hot = {
      size        = "1g"
      zone_count  = 1
      autoscaling = {}
    }
  }

  kibana = {
    config = {
      docker_image = "docker.elastic.co/cloud-ci/kibana:7.15.0-SNAPSHOT"
    }
  }

  apm = {
    config = {
      docker_image = "docker.elastic.co/cloud-ci/apm:7.15.0-SNAPSHOT"
    }
  }

  enterprise_search = {
    config = {
      docker_image = "docker.elastic.co/cloud-ci/enterprise-search:7.15.0-SNAPSHOT"
    }

    zone_count = 1
  }
}
