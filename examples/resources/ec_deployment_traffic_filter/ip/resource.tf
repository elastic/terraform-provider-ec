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

  traffic_filter = [
    ec_deployment_traffic_filter.example.id
  ]

  # Use the deployment template defaults
  elasticsearch = {
    hot = {
      autoscaling = {}
    }
  }

  kibana = {}
}

resource "ec_deployment_traffic_filter" "example" {
  name   = "my traffic filter name"
  region = "us-east-1"
  type   = "ip"

  rule {
    source = "0.0.0.0/0"
  }
}
