variable "name" {
  type = string
}

variable "region" {
  type = string
}

resource "ec_deployment_traffic_filter" "basic" {
  name   = var.name
  region = var.region
  type   = "ip"

  rule {
    source = "0.0.0.0/0"
  }
}
