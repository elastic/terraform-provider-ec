resource "ec_serverless_traffic_filter" "example" {
  name        = "my-serverless-traffic-filter"
  region      = "aws-us-east-1"
  type        = "ip"
  description = "Allow traffic from the office network"

  include_by_default = false

  rules {
    source      = "203.0.113.0/24"
    description = "Office egress"
  }
}
