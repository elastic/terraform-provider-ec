resource "ec_deployment" "basic" {
  name    = "%s"
  region  = "%s"
  version = "%s"

  # TODO: Make this template ID dependent on the region.
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch {
    topology {
      config {
        user_settings_yaml = "action.auto_create_index: true"
      }
      instance_configuration_id = "aws.data.highio.i3"
      memory_per_node           = "1g"
    }
  }

  kibana {
    topology {
      config {
        user_settings_yaml = "csp.warnLegacyBrowsers: true"
      }
      instance_configuration_id = "aws.kibana.r5d"
    }
  }

  apm {
    topology {
      config {
        debug_enabled = true
      }
      instance_configuration_id = "aws.apm.r5d"
    }
  }

  enterprise_search {
    topology {
      config {
        user_settings_yaml = "ent_search.login_assistance_message: somemessage"
      }
      instance_configuration_id = "aws.enterprisesearch.m5d"
    }
  }
}