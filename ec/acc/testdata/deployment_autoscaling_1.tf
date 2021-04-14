data "ec_stack" "autoscaling" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "autoscaling" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.autoscaling.version
  deployment_template_id = "%s"

  elasticsearch {
    autoscale = "true"
    
    topology {
      id         = "hot_content"
      size       = "1g"
      zone_count = 1
      autoscaling {
        max_size = "8g"
      }
    }
    topology {
      id         = "ml"
      size       = "1g"
      zone_count = 1
      autoscaling {
        min_size = "1g"
        max_size = "4g"
      }
    }
    topology {
      id         = "warm"
      size       = "2g"
      zone_count = 1
      autoscaling {
        max_size = "15g"
      }
    }
  }
}