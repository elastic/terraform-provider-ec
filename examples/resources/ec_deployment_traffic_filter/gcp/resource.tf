locals {
  region = asia-east1
}

data "ec_stack" "latest" {
  version_regex = "latest"
  region        = local.region
}

# Create an Elastic Cloud deployment
resource "ec_deployment" "example_minimal" {
  # Optional name.
  name = "my_example_deployment"

  # Mandatory fields
  region                 = local.region
  version                = data.ec_stack.latest.version
  deployment_template_id = "gcp-storage-optimized"

  traffic_filter = [
    ec_deployment_traffic_filter.gcp_psc.id
  ]

  # Use the deployment template defaults
  elasticsearch = {
    hot = {
      autoscaling = {}
    }
  }

  kibana = {}
}

resource "ec_deployment_traffic_filter" "gcp_psc" {
  name   = "my traffic filter name"
  region = local.region
  type   = "gcp_private_service_connect_endpoint"

  rule {
    source = "18446744072646845332"
  }
}
