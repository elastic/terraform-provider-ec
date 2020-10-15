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
    topology {
      config {
        user_settings_yaml = "action.auto_create_index: true"
      }
      size = "1g"
    }
  }

  kibana {
    topology {
      config {
        user_settings_yaml = "csp.warnLegacyBrowsers: true"
      }
    }
  }

  apm {
    topology {
      config {
        debug_enabled = true
        user_settings_json = jsonencode({"apm-server.rum.enabled"= true})
      }
    }
  }

  enterprise_search {
    topology {
      config {
        user_settings_yaml = "ent_search.login_assistance_message: somemessage"
      }
    }
  }
}