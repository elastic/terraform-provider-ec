data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "autoscale_ml" {
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
    ml = {
      autoscaling = {
        autoscale = true
      }
    }
  }
}

