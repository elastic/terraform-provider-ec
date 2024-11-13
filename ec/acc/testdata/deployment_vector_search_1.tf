data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "vector_search" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch = {
    hot = {
      autoscaling = {}
    }

    "remote_cluster" = [for source_vector_search in ec_deployment.source_vector_search :
      {
        deployment_id = source_vector_search.id
        alias         = source_vector_search.name
      }
    ]
  }
}

resource "ec_deployment" "source_vector_search" {
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
