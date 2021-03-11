data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "basic" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch {
    config {
      user_settings_yaml = "action.auto_create_index: true"
    }
    topology {
      id   = "hot_content"
      size = "1g"
    }
  }

  kibana {
    config {
      user_settings_yaml = "csp.warnLegacyBrowsers: true"
    }
    topology {
      instance_configuration_id = "%s"
    }
  }

  apm {
    config {
      debug_enabled      = true
      user_settings_json = jsonencode({ "apm-server.rum.enabled" = true })
    }
    topology {
      instance_configuration_id = "%s"
    }
  }

  enterprise_search {
    config {
      user_settings_yaml = "ent_search.login_assistance_message: somemessage"
    }
    topology {
      instance_configuration_id = "%s"
    }
  }
}