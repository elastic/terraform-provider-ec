data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
}

# Create an Elastic Cloud deployment
resource "ec_deployment" "example_keystore" {
  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch = {
    hot = {
      autoscaling = {}
    }
  }
}

# Create the keystore secret entry
resource "ec_deployment_elasticsearch_keystore" "gcs_credential" {
  deployment_id = ec_deployment.example_keystore.id
  setting_name  = "gcs.client.default.credentials_file"
  value         = file("service-account-key.json")
  as_file       = true
}
