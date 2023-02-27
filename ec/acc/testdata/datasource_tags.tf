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
    "foo"     = "bar"
    "bar"     = "baz"
    "test_id" = "%s"
  }

  elasticsearch = {
    hot = {
      size        = "1g"
      zone_count  = 1
      autoscaling = {}
    }
  }
}

data "ec_deployment" "tagdata" {
  id = ec_deployment.tags.id
}

data "ec_deployments" "tagfilter" {
  tags = {
    "test_id" = "%s"
  }
}