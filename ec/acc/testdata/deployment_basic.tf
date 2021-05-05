data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "basic" {
  alias                  = "%s"
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch {
    topology {
      id   = "hot_content"
      size = "1g"
    }
  }

  kibana {
    topology {
      instance_configuration_id = "%s"
    }
  }

  apm {
    topology {
      instance_configuration_id = "%s"
    }
  }

  enterprise_search {
    topology {
      instance_configuration_id = "%s"
    }
  }
}