resource "ec_deployment_traffic_filter" "basic" {
  name   = "%s"
  region = "%s"
  type   = "ip"

  rule {
    source = "0.0.0.0/0"
  }

  rule {
    source = "1.1.1.0/24"
  }
}
