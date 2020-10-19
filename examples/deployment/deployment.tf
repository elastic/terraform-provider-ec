terraform {
  required_version = ">= 0.12.29"

  required_providers {
    ec = {
      source = "elastic/ec"
    }
  }
}

provider "ec" {}

# Create an Elastic Cloud deployment
resource "ec_deployment" "example_minimal" {
  # Optional name.
  name = "my_example_deployment"

  # Mandatory fields
  region                 = "us-east-1"
  version                = "7.9.2"
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch {}

  kibana {}

  observability {
    deployment_id = "f759065e5e64e9f3546f6c44f2743893"
    metrics = "false"
  }
}
