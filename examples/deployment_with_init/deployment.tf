# Retrieve the latest stack pack version
data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
}

# Create an Elastic Cloud deployment
resource "ec_deployment" "example_minimal" {
  # Optional name.
  name = "my_example_deployment"

  # Mandatory fields
  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"
  traffic_filter         = [ec_deployment_traffic_filter.allow_all.id]
  elasticsearch = {
    config = {}
    hot = {
      size        = "8g"
      autoscaling = {}
    }
  }

  kibana = {
    size = "1g"
  }
}

resource "ec_deployment_traffic_filter" "allow_all" {
  name   = "Allow all ip addresses"
  region = "us-east-1"
  type   = "ip"

  rule {
    source = "0.0.0.0/0"
  }
}

output "elasticsearch_https_endpoint" {
  value = ec_deployment.example_minimal.elasticsearch[0].https_endpoint
}

output "elasticsearch_username" {
  value = ec_deployment.example_minimal.elasticsearch_username
}

output "elasticsearch_password" {
  value = ec_deployment.example_minimal.elasticsearch_password
}

output "elasticsearch_cloud_id" {
  value = ec_deployment.example_minimal.elasticsearch[0].cloud_id
}
