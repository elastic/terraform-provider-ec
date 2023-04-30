data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
}

resource "ec_deployment" "example_observability" {
  # Optional name.
  name = "my_example_deployment"

  # Mandatory fields
  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch = {
    hot = {
      autoscaling = {}
    }
  }

  kibana = {}

  # Optional observability settings
  observability = {
    deployment_id = ec_deployment.example_minimal.id
  }

  tags = {
    "monitoring" = "source"
  }
}
