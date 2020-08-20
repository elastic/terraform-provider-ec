resource "ec_deployment" "enterprise_search" {
  name    = "%s"
  region  = "%s"
  version = "%s"

  # TODO: Make this template ID dependent on the region.
  deployment_template_id = "aws-enterprise-search-dedicated-v2"

  elasticsearch {
    topology {
      instance_configuration_id = "aws.data.highcpu.m5d"
      memory_per_node           = "1g"
    }
  }

  kibana {
    topology {
      instance_configuration_id = "aws.kibana.r5d"
    }
  }

  enterprise_search {
    config {
      user_settings_yaml = "ent_search.auth.source: standard"
    }
    topology {
      instance_configuration_id = "aws.enterprisesearch.m5d"
    }
  }
}