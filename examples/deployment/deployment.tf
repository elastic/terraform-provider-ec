terraform {
  # The Elastic Cloud provider is supported from ">=0.12"
  # Version later than 0.12.29 is required for this terraform block to work.
  required_version = ">= 0.12.29"

  required_providers {
    ec = {
      source  = "elastic/ec"
      version = "0.3.0"
    }
  }
}

provider "ec" {}

# Retrieve the latest stack pack version
data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
}

# Create an Elastic Cloud deployment
resource "ec_deployment" "example_minimal" {
  # Optional name.
  name = "my_example_deployment"

  region                 = "gcp-europe-west4"
  version                = "7.13.2"
  deployment_template_id = "gcp-hot-warm"

  elasticsearch {
    config {
      user_settings_yaml = "# test"
    }
  }
}

resource "ec_deployment_elasticsearch_keystore" "test" {
  count         = 1
  deployment_id = ec_deployment.example_minimal.id
  setting_name  = "xpack.notification.slack.account.hello.secure_url"
  value         = "hello"
}

resource "ec_deployment_elasticsearch_keystore" "world" {
  count         = 1
  deployment_id = ec_deployment.example_minimal.id
  setting_name  = "xpack.notification.slack.account.world.secure_url"
  value         = "world"
}

resource "ec_deployment_elasticsearch_keystore" "itme" {
  count         = 1
  deployment_id = ec_deployment.example_minimal.id
  setting_name  = "xpack.notification.slack.account.yay.secure_url"
  value         = "woop"
}
