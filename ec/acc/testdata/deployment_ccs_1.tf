data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "ccs" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch = {
    hot = {
      autoscaling = {}
    }

    "remote_cluster" = [for source_css in ec_deployment.source_ccs :
      {
        deployment_id = source_css.id
        alias         = source_css.name
      }
    ]
  }
}

resource "ec_deployment" "source_ccs" {
  count                  = 3
  name                   = "%s-${count.index}"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch = {
    hot = {
      zone_count  = 1
      size        = "1g"
      autoscaling = {}
    }
  }
}
