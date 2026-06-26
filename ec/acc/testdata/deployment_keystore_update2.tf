data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "test" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch = {
    hot = {
      size        = "1g"
      autoscaling = {}
    }

    ml = {
      autoscaling = {}
    }
  }

  kibana = {
    zone_count = 1
  }

}

resource "ec_deployment_elasticsearch_keystore" "test" {
  deployment_id = ec_deployment.test.id
  setting_name  = "xpack.notification.slack.account.monitoring.secure_url"
  value         = "secret-2"
}
