data "ec_stack" "pre_node_roles" {
  version_regex = "7.??.?"
  region        = "%s"
}

resource "ec_deployment" "pre_nr" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.pre_node_roles.version
  deployment_template_id = "%s"

  elasticsearch = {
    hot = {
      size        = "1g"
      zone_count  = 1
      autoscaling = {}
    }
  }
}