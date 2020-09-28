resource "ec_deployment" "basic" {
  name    = "%s"
  region  = "%s"
  version = "%s"

  # TODO: Make this template ID dependent on the region.
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch {
    config {
      user_settings_yaml = "action.auto_create_index: true"
    }
    topology {
      instance_configuration_id = "aws.data.highio.i3"
      memory_per_node           = "1g"
    }
  }

  kibana {
    config {
      user_settings_yaml = "csp.warnLegacyBrowsers: true"
    }
    topology {
      instance_configuration_id = "aws.kibana.r5d"
    }
  }

  apm {
    config {
      debug_enabled = true
    }
    topology {
      instance_configuration_id = "aws.apm.r5d"
    }
  }

  enterprise_search {
    config {
      user_settings_yaml = "ent_search.login_assistance_message: somemessage"
    }
    topology {
      instance_configuration_id = "aws.enterprisesearch.m5d"
    }
  }
}