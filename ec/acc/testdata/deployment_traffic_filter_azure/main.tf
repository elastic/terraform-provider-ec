variable "name" {
  type = string
}

variable "region" {
  type = string
}

resource "ec_deployment_traffic_filter" "azure" {
  name   = var.name
  region = var.region
  type   = "azure_private_endpoint"

  rule {
    azure_endpoint_name = "my-azure-pl"
    azure_endpoint_guid = "78c64959-fd88-41cc-81ac-1cfcdb1ac32e"
  }
}
