data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "ccs" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch {
    remote_cluster {
      deployment_id = ec_deployment.source_ccs.id
      alias         = "my_source_ccs"
    }
  }
}

resource "ec_deployment" "source_ccs" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch {
    topology {
      zone_count = 1
      size       = "1g"
    }
  }
}