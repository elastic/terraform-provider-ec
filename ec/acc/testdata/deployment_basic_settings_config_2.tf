data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "basic" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch = {
    config = {
      user_settings_yaml = "action.auto_create_index: true"
    }
    hot = {
      size        = "1g"
      autoscaling = {}
    }
  }

  kibana = {
    config = {
      user_settings_yaml = "csp.warnLegacyBrowsers: true"
    }

    instance_configuration_id = "%s"
  }

  apm = {
    config = {
      debug_enabled      = true
      user_settings_json = jsonencode({ "apm-server.rum.enabled" = true })
    }

    instance_configuration_id = "%s"
  }

  enterprise_search = {
    config = {
      user_settings_yaml = "# comment"
    }

    instance_configuration_id = "%s"
  }
}