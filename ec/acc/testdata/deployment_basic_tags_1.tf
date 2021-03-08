data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "tags" {
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

  tags = {
    owner       = "elastic"
    cost-center = "rnd"
  }
}