variable "name" {
  type = string
}

variable "region" {
  type = string
}

variable "google_dns_rules" {
  type = list(string)
  default = [
    "8.8.8.8/24", "8.8.4.4/24", "8.8.8.9/24", "8.8.4.10/24", "8.8.8.11/24", "8.8.4.12/24", "8.8.8.13/24", "8.8.4.14/24",
    "9.8.8.8/24", "10.8.4.4/24", "11.8.8.9/24", "12.8.4.10/24", "13.8.8.11/24", "14.8.4.12/24", "15.8.8.13/24", "16.8.4.14/24",
  ]
}

resource "ec_deployment_traffic_filter" "basic" {
  name   = var.name
  region = var.region
  type   = "ip"

  dynamic "rule" {
    for_each = var.google_dns_rules
    content {
      source = rule.value
    }
  }
}
