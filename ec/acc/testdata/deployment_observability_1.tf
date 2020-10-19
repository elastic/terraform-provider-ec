data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "basic" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch {
    topology {
      size       = "1g"
      zone_count = 1
    }
  }
}

resource "ec_deployment" "observability" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch {
    topology {
      size       = "1g"
      zone_count = 1
    }
  }

  observability {
    deployment_id = ec_deployment.basic.id
  }
}