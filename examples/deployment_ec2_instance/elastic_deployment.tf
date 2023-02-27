# Retrieve the latest stack pack version
data "ec_stack" "latest" {
  version_regex = "latest"
  region        = var.region
}

# Create an Elastic Cloud deployment
resource "ec_deployment" "deployment" {
  # Optional name.
  name = "elasticsearch_deployment"

  # Mandatory fields.
  region                 = var.region
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"
  traffic_filter         = [ec_deployment_traffic_filter.allow_my_instance.id]

  # Note the deployment will contain Elasticsearch and Kibana resources with default configurations.
  elasticsearch = {
    config = {}
    hot = {
      autoscaling = {}
    }
  }

  kibana = {}
}

# Create a traffic filter to allow the instance's public IP address to access our deployment.
# This can also be done using a VPC private link connection.
resource "ec_deployment_traffic_filter" "allow_my_instance" {
  name   = format("Allow %s", aws_instance.web.id)
  region = var.region
  type   = "ip"

  rule {
    # Render the IP address with an additional /32 for full CIDR address.
    source = format("%s/32", aws_instance.web.public_ip)
  }
}
