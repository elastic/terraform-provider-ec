data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "memory_optimized" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch {
    topology {
      id   = "hot_content"
      size = "2g"
    }
  }

  kibana {}

  apm {}
}