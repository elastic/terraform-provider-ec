data "ec_deployment" "example" {
  id = "320b7b540dfc967a7a649c18e2fce4ed"
}

resource "ec_deployment_traffic_filter" "example" {
  name   = "my traffic filter name"
  region = "us-east-1"
  type   = "ip"

  rule {
    source = "0.0.0.0/0"
  }
}

resource "ec_deployment_traffic_filter_association" "example" {
  traffic_filter_id = ec_deployment_traffic_filter.example.id
  deployment_id     = ec_deployment.example.id
}
