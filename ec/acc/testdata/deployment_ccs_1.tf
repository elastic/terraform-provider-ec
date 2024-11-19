data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "general_purpose" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch = {
    hot = {
      autoscaling = {}
    }

    "remote_cluster" = [for source_storage_optimized in ec_deployment.source_storage_optimized :
      {
        deployment_id = source_storage_optimized.id
        alias         = source_storage_optimized.name
      }
    ]
  }
}

resource "ec_deployment" "source_storage_optimized" {
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
