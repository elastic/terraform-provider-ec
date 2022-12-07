data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "tf_assoc" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch = {
    hot = {
      size        = "1g"
      autoscaling = {}
    }
  }

  kibana = {}
}

resource "ec_deployment_traffic_filter" "tf_assoc_second" {
  name   = "%s"
  region = "%s"
  type   = "ip"

  rule {
    source = "0.0.0.0/0"
  }
}

resource "ec_deployment_traffic_filter_association" "tf_assoc" {
  traffic_filter_id = ec_deployment_traffic_filter.tf_assoc_second.id
  deployment_id     = ec_deployment.tf_assoc.id
}
