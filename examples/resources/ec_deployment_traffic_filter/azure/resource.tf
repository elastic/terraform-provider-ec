locals {
  region = azure-australiaeast
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
  deployment_template_id = "azure-io-optimized-v3"

  traffic_filter = [
    ec_deployment_traffic_filter.azure.id
  ]

  # Use the deployment template defaults
  elasticsearch = {
    hot = {
      autoscaling = {}
    }
  }

  kibana = {}
}

resource "ec_deployment_traffic_filter" "azure" {
  name   = "my traffic filter name"
  region = local.region
  type   = "azure_private_endpoint"

  rule {
    azure_endpoint_name = "my-azure-pl"
    azure_endpoint_guid = "78c64959-fd88-41cc-81ac-1cfcdb1ac32e"
  }
}
