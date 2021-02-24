data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "tags" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"
  tags = {
    "foo" = "bar"
    "bar" = "baz"
  }

  elasticsearch {
    topology {
      size       = "1g"
      zone_count = 1
    }
  }
}

data "ec_deployment" "tags" {
  id = ec_deployment.tags.id
}