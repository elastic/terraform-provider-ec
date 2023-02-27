data "ec_stack" "keystore" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "keystore" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.keystore.version
  deployment_template_id = "%s"

  elasticsearch = {
    hot = {
      size        = "1g"
      zone_count  = 1
      autoscaling = {}
    }
  }
}

resource "ec_deployment_elasticsearch_keystore" "test" {
  deployment_id = ec_deployment.keystore.id
  setting_name  = "xpack.notification.slack.account.hello.secure_urla"
  value         = "hello2u"
}

resource "ec_deployment_elasticsearch_keystore" "gcs_creds" {
  deployment_id = ec_deployment.keystore.id
  setting_name  = "gcs.client.secondary.credentials_file"
  value         = file("testdata/deployment_elasticsearch_keystore_creds.json")
}
