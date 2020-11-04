# Create an Elastic Cloud deployment
resource "ec_deployment" "deployment" {
  # Optional name.
  name = "elasticsearch_deployment"

  # Mandatory fields
  region                 = var.region
  version                = var.deployment_version
  deployment_template_id = "aws-io-optimized-v2"
  traffic_filter         = [ec_deployment_traffic_filter.allow_my_instance.id]

  # Noting the deploymnet will contain elasticsearch and kibana with default configurations
  elasticsearch {}
  kibana {}
}

# Create a traffic filter to allow the instance's public IP address to access our deployment.
# This can also be done using a VPC private link connection.
resource "ec_deployment_traffic_filter" "allow_my_instance" {
  name   = format("Allow %s", aws_instance.inst.id)
  region = var.region
  type   = "ip"

  rule {
    # Render the IP address with an additional /32 for full CIDR address.
    source = format("%s/32", aws_instance.inst.public_ip)
  }
}
