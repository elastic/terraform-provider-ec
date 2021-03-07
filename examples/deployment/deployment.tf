terraform {
  required_version = ">= 0.12"

  required_providers {
    ec = {
      source  = "elastic/ec"
      version = ">=0.1.0-beta"
    }
  }
}

provider "ec" {
  apikey="TmxIdURIZ0JxVld3ME5DWTdqVUM6WnhTWEZoWXpRRmVMMXh6WHQ0U2RIQQ=="
}

# Create an Elastic Cloud deployment
resource "ec_deployment" "example_minimal" {
  # Optional name.
  name = "my_example_deployment"

  # Mandatory fields
  region                 = "us-east-1"
  version                = "7.9.2"
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch {
    config {
	user_settings_yaml = file("./es_settings.yaml")
    }
  }

  kibana {}
}
