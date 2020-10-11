data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "defaults" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch {}

  kibana {
    topology {
      size = "2g"
    }
  }

  apm {
    topology {
      size = "1g"
    }
  }

  enterprise_search {
    topology {
      zone_count = 1
    }
  }
}