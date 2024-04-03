data "ec_deployment_templates" "example" {
  region = "us-east-1"
}

resource "ec_deployment" "my_deployment" {
  name                   = "My Deployment"
  version                = "8.12.2"
  region                 = data.ec_deployment_templates.all_templates.region
  deployment_template_id = data.ec_deployment_templates.all_templates.templates.0.id

  elasticsearch = {
    hot = {
      autoscaling = {}
    }
  }

  kibana = {}
}